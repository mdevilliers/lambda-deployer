resource "aws_iam_role" "deployer" {
  name = "${var.application}_${var.environment}_deployer_role"

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

data "aws_caller_identity" "current" {}

resource "aws_iam_role_policy" "deployer" {
  name = "${var.application}_${var.environment}_deployer_identity"
  role = "${aws_iam_role.deployer.id}"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole",
        "lambda:CreateAlias",
        "lambda:CreateFunction",
        "lambda:GetAlias",
        "lambda:GetFunction",
        "lambda:UpdateAlias",
        "lambda:UpdateFunctionCode",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents",
        "s3:GetObject",
        "s3:HeadObject"
      ],
      "Resource": [
        "arn:aws:logs:*:*:*",
        "${var.s3_bucket_arn}/*",
        "arn:aws:lambda:*:${data.aws_caller_identity.current.account_id}:*",
        "arn:aws:iam::${data.aws_caller_identity.current.account_id}:*"
      ]
    }
  ]
}
EOF
}

resource "aws_lambda_function" "deployer" {
  function_name    = "${var.application}_${var.environment}_deployer"
  handler          = "handler.Handle"
  runtime          = "python2.7"
  filename         = "${var.deployer_filepath}"
  source_code_hash = "${base64sha256(file(var.deployer_filepath))}"
  role             = "${aws_iam_role.deployer.arn}"
  timeout          = 120

  environment {
    variables = {
      DEPLOYER_FUNCTION_ROLE_ARN = "${var.function_role_arn}"
    }
  }
}

resource "aws_lambda_permission" "allow_s3" {
  statement_id  = "AllowExecutionFromS3Bucket"
  action        = "lambda:InvokeFunction"
  function_name = "${aws_lambda_function.deployer.function_name}"
  principal     = "s3.amazonaws.com"
  source_arn    = "${var.s3_bucket_arn}"
}

resource "aws_s3_bucket_notification" "deployment" {
  bucket = "${var.s3_bucket_id}"

  lambda_function {
    lambda_function_arn = "${aws_lambda_function.deployer.arn}"
    events              = ["s3:ObjectCreated:*"]
  }
}
