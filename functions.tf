resource "aws_iam_role" "iot_button_metrics" {
  name = "iot_button_metrics"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "iot_button_metrics" {
  name = "iot_button_metrics"
  role = "${aws_iam_role.iot_button_metrics.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
       "Effect": "Allow",
       "Action": ["s3:ListBucket"],
       "Resource": ["${aws_s3_bucket.metrics.arn}"]
     },
     {
       "Effect": "Allow",
       "Action": [
         "s3:PutObject",
         "s3:GetObject",
         "s3:DeleteObject"
       ],
       "Resource": ["${aws_s3_bucket.metrics.arn}/*"]
     },
     {
       "Effect": "Allow",
       "Action": [
         "logs:*"        
       ],
       "Resource": "arn:aws:logs:*:*:*"
     }
  ]
}
EOF
}

resource "aws_lambda_function" "send_used_metric" {
  filename         = "bin/send_used_metric.zip"
  function_name    = "send_used_metric"
  role             = "${aws_iam_role.iot_button_metrics.arn}"
  handler          = "bin/send_used_metric"
  source_code_hash = "${base64sha256(file("bin/send_used_metric.zip"))}"
  runtime          = "go1.x"

  environment {
    variables = {
      BUCKET                  = "${var.tags}-${var.provier_name}"
      LASTMODIFIED_KEY_PRIFIX = "${var.lastmodified_key_prifix}"
    }
  }

  tags {
    Name        = "${var.tags}"
    Environment = "Production"
  }
}

resource "aws_lambda_function" "send_scheduled_metric" {
  filename         = "bin/send_scheduled_metric.zip"
  function_name    = "send_scheduled_metric"
  role             = "${aws_iam_role.iot_button_metrics.arn}"
  handler          = "bin/send_scheduled_metric"
  source_code_hash = "${base64sha256(file("bin/send_scheduled_metric.zip"))}"
  runtime          = "go1.x"

  environment {
    variables = {
      BUCKET                  = "${var.tags}-${var.provier_name}"
      METRICS_KEY_PRIFIX      = "${var.metrics_key_prifix}"
      LASTMODIFIED_KEY_PRIFIX = "${var.lastmodified_key_prifix}"
      TIMEOUT                 = "${var.timeout}"
      BUTTON_COUNT            = "${var.button_count}"
      BUTTON_PREFIX           = "${var.button_prefix}"
    }
  }

  tags {
    Name        = "${var.tags}"
    Environment = "Production"
  }
}

resource "aws_cloudwatch_event_rule" "every_month" {
  name                = "every_minute"
  description         = "Fires every minute"
  schedule_expression = "cron(* * * * ? *)"
}

resource "aws_cloudwatch_event_target" "output_report_every_month" {
  rule      = "${aws_cloudwatch_event_rule.every_month.name}"
  target_id = "output_report"
  arn       = "${aws_lambda_function.send_scheduled_metric.arn}"
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_output_report" {
  statement_id  = "AllowExecutionFromCloudWatch"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.send_scheduled_metric.function_name}"
  principal     = "events.amazonaws.com"
  source_arn    = "${aws_cloudwatch_event_rule.every_month.arn}"
}

resource "aws_lambda_function" "slack_command" {
  filename         = "bin/slack_command.zip"
  function_name    = "slack_command"
  role             = "${aws_iam_role.iot_button_metrics.arn}"
  handler          = "bin/slack_command"
  source_code_hash = "${base64sha256(file("bin/slack_command.zip"))}"
  runtime          = "go1.x"

  environment {
    variables = {
      BUCKET                  = "${var.tags}-${var.provier_name}"
      LASTMODIFIED_KEY_PRIFIX = "${var.lastmodified_key_prifix}"
      TIMEOUT                 = "${var.timeout}"
      BUTTON_COUNT            = "${var.button_count}"
      BUTTON_PREFIX           = "${var.button_prefix}"
      SLACK_VERIFIED_TOKEN    = "${var.slack_verified_token}"
    }
  }

  tags {
    Name        = "${var.tags}"
    Environment = "Production"
  }
}

resource "aws_api_gateway_rest_api" "slack_command" {
  name = "SlackCommand"
}

resource "aws_api_gateway_resource" "slack_command" {
  rest_api_id = "${aws_api_gateway_rest_api.slack_command.id}"
  parent_id   = "${aws_api_gateway_rest_api.slack_command.root_resource_id}"
  path_part   = "slack_command"
}

resource "aws_api_gateway_method" "slack_command" {
  rest_api_id   = "${aws_api_gateway_rest_api.slack_command.id}"
  resource_id   = "${aws_api_gateway_resource.slack_command.id}"
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "slack_command_post" {
  rest_api_id = "${aws_api_gateway_rest_api.slack_command.id}"
  resource_id = "${aws_api_gateway_resource.slack_command.id}"
  http_method = "${aws_api_gateway_method.slack_command.http_method}"
  status_code = "200"

  response_parameters {
    "method.response.header.Access-Control-Allow-Origin" = true
  }

  depends_on = ["aws_api_gateway_method.slack_command"]
}

resource "aws_api_gateway_integration" "slack_command_post" {
  rest_api_id = "${aws_api_gateway_rest_api.slack_command.id}"
  resource_id = "${aws_api_gateway_method.slack_command.resource_id}"
  http_method = "${aws_api_gateway_method.slack_command.http_method}"

  integration_http_method = "POST"
  type                    = "AWS"
  uri                     = "${aws_lambda_function.slack_command.invoke_arn}"

  passthrough_behavior = "WHEN_NO_TEMPLATES"

  # Transforms the incoming XML request to JSON
  request_templates {
    "application/x-www-form-urlencoded" = <<EOF
## convert HTML POST data or HTTP GET query string to JSON

## get the raw post data from the AWS built-in variable and give it a nicer name
#if ($context.httpMethod == "POST")
 #set($rawAPIData = $input.path('$'))
 ## escape any quotes
 #set($rawAPIData = $rawAPIData.replace('"', '\"'))
#elseif ($context.httpMethod == "GET")
 #set($rawAPIData = $input.params().querystring)
 #set($rawAPIData = $rawAPIData.toString())
 #set($rawAPIDataLength = $rawAPIData.length() - 1)
 #set($rawAPIData = $rawAPIData.substring(1, $rawAPIDataLength))
 #set($rawAPIData = $rawAPIData.replace(", ", "&"))
#else
 #set($rawAPIData = "")
#end

## first we get the number of "&" in the string, this tells us if there is more than one key value pair
#set($countAmpersands = $rawAPIData.length() - $rawAPIData.replace("&", "").length())

## if there are no "&" at all then we have only one key value pair.
## we append an ampersand to the string so that we can tokenise it the same way as multiple kv pairs.
## the "empty" kv pair to the right of the ampersand will be ignored anyway.
#if ($countAmpersands == 0)
 #set($rawPostData = $rawAPIData + "&")
#end

## now we tokenise using the ampersand(s)
#set($tokenisedAmpersand = $rawAPIData.split("&"))

## we set up a variable to hold the valid key value pairs
#set($tokenisedEquals = [])

## now we set up a loop to find the valid key value pairs, which must contain only one "="
#foreach( $kvPair in $tokenisedAmpersand )
 #set($countEquals = $kvPair.length() - $kvPair.replace("=", "").length())
 #if ($countEquals == 1)
  #set($kvTokenised = $kvPair.split("="))
  #if ($kvTokenised[0].length() > 0)
   ## we found a valid key value pair. add it to the list.
   #set($devNull = $tokenisedEquals.add($kvPair))
  #end
 #end
#end

## next we set up our loop inside the output structure "{" and "}"
{
#foreach( $kvPair in $tokenisedEquals )
  ## finally we output the JSON for this pair and append a comma if this isn't the last pair
  #set($kvTokenised = $kvPair.split("="))
  #if($kvTokenised.size() == 2 && $kvTokenised[1].length() > 0)
    #set($kvValue = $kvTokenised[1])
  #else
    #set($kvValue = "")
  #end
  #if( $foreach.hasNext )
    #set($itemDelimiter = ",")
  #else
    #set($itemDelimiter = "")
  #end
 "$kvTokenised[0]" : "$util.urlDecode($kvValue)"$itemDelimiter
#end
}
EOF
  }
}

resource "aws_api_gateway_method" "options_method" {
  rest_api_id   = "${aws_api_gateway_rest_api.slack_command.id}"
  resource_id   = "${aws_api_gateway_resource.slack_command.id}"
  http_method   = "OPTIONS"
  authorization = "NONE"
}

resource "aws_api_gateway_method_response" "options_200" {
  rest_api_id = "${aws_api_gateway_rest_api.slack_command.id}"
  resource_id = "${aws_api_gateway_resource.slack_command.id}"
  http_method = "${aws_api_gateway_method.options_method.http_method}"
  status_code = "200"

  response_models {
    "application/json" = "Empty"
  }

  response_parameters {
    "method.response.header.Access-Control-Allow-Headers" = true
    "method.response.header.Access-Control-Allow-Methods" = true
    "method.response.header.Access-Control-Allow-Origin"  = true
  }

  depends_on = ["aws_api_gateway_method.options_method"]
}

resource "aws_api_gateway_integration" "options_integration" {
  rest_api_id = "${aws_api_gateway_rest_api.slack_command.id}"
  resource_id = "${aws_api_gateway_resource.slack_command.id}"
  http_method = "${aws_api_gateway_method.options_method.http_method}"
  type        = "MOCK"
  depends_on  = ["aws_api_gateway_method.options_method"]
}

resource "aws_api_gateway_integration_response" "options_integration_response" {
  rest_api_id = "${aws_api_gateway_rest_api.slack_command.id}"
  resource_id = "${aws_api_gateway_resource.slack_command.id}"
  http_method = "${aws_api_gateway_method.options_method.http_method}"
  status_code = "${aws_api_gateway_method_response.options_200.status_code}"

  response_parameters = {
    "method.response.header.Access-Control-Allow-Headers" = "'Content-Type,X-Amz-Date,Authorization,X-Api-Key,X-Amz-Security-Token'"
    "method.response.header.Access-Control-Allow-Methods" = "'GET,OPTIONS,POST,PUT'"
    "method.response.header.Access-Control-Allow-Origin"  = "'*'"
  }

  depends_on = ["aws_api_gateway_method_response.options_200"]
}

resource "aws_api_gateway_deployment" "slack_command" {
  depends_on = [
    "aws_api_gateway_integration.slack_command_post",
    "aws_api_gateway_integration.options_integration",
  ]

  rest_api_id = "${aws_api_gateway_rest_api.slack_command.id}"
  stage_name  = "prod"
}
