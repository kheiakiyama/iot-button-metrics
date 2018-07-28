# iot-button-metrics
iot-button-metrics is used for collecting metrics of toilet

![structure](https://raw.githubusercontent.com/kheiakiyama/iot-button-metrics/images/images/20180728_AWS-IoT-Button_key.png)

Output csv metrics to s3 bucket like below.

```
"2018/07/28 12:12:30",0
"2018/07/28 12:13:29",0
"2018/07/28 12:14:30",1
"2018/07/28 12:15:30",1
"2018/07/28 12:16:30",1
"2018/07/28 12:17:29",1
"2018/07/28 12:18:30",1
"2018/07/28 12:19:30",0
"2018/07/28 12:20:30",0
...
```

If you push [AWS IoT Button](https://aws.amazon.com/iotbutton), save metric `"1"` until 300 seconds.  
After that, save metric `"0"`.

## Get started
1. `git clone git@github.com:kheiakiyama/iot-button-metrics.git`
2. Copy `terraform.tfvars.example` to `terraform.tfvars` and edit variables
3. `sh build_function.sh`
4. `terraform apply`
5. Set [AWS IoT Button](https://aws.amazon.com/iotbutton) call `send_used_metric`.  
Each buttons, set placement attribute `DEVICE=button1`,`DEVICE=button2`,`DEVICE=button3`...  
