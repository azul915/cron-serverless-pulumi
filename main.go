package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/cloudwatch"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v4/go/aws/lambda"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {

		err := launch(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}

func launch(ctx *pulumi.Context) error {

	// Create an IAM Policy for Lambda AssumeRole
	lambdaRole, err := iam.NewRole(ctx, "lambda-role", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.String(`{
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
		}`),
	})
	if err != nil {
		return err
	}

	// policy, err := iam.NewPolicy(ctx, "lambda-policy", &iam.PolicyArgs{
	// 	Policy: pulumi.String(`{
	// 				"Version": "2012-10-17",
	// 				"Statement": [
	// 					{
	// 						"Effect": "Allow",
	// 						"Action": [
	// 							"logs:CreateLogGroup",
	// 							"logs:CreateLogStream",
	// 							"logs:PutLogEvents"
	// 						],
	// 						"Resource": "*"
	// 					}
	// 				]
	// 	}`),
	// })

	// Create an Lambda Function
	lambdaFunction, err := lambda.NewFunction(ctx, "lambda-function", &lambda.FunctionArgs{
		Description: pulumi.String("lambda function desicription"),
		Runtime:     pulumi.String("go1.x"),
		Name:        pulumi.String("lambda-function-name"),
		Handler:     pulumi.String("entrypoint"),
		Code:        pulumi.NewFileArchive("./sample.zip"),
		Role:        lambdaRole.Arn,
		Environment: &lambda.FunctionEnvironmentArgs{
			Variables: pulumi.StringMap{
				"FOO": pulumi.String("BAR"),
			},
		},
	})
	if err != nil {
		return err
	}

	// Create an CloudWatch Events
	eventRule, err := cloudwatch.NewEventRule(ctx, "rule", &cloudwatch.EventRuleArgs{
		Name:               pulumi.Sprintf("%s-kick-rule", lambdaFunction.Name),
		Description:        pulumi.String("cloudwatchevents role description"),
		ScheduleExpression: pulumi.String("cron(00 8,20 * * ? *)"),
	})
	if err != nil {
		return err
	}

	_, err = lambda.NewPermission(ctx, "lambda-permission-cloudwatchevents", &lambda.PermissionArgs{
		Action:    pulumi.String("lambda:InvokeFunction"),
		Principal: pulumi.String("events.amazonaws.com"),
		SourceArn: eventRule.Arn,
		Function:  lambdaFunction.Name,
	})
	if err != nil {
		return err
	}

	_, err = cloudwatch.NewEventTarget(ctx, "lambda", &cloudwatch.EventTargetArgs{
		Rule: eventRule.Name,
		Arn:  lambdaFunction.Arn,
	})
	if err != nil {
		return err
	}

	_, err = cloudwatch.NewLogGroup(ctx, "lgtm-log", &cloudwatch.LogGroupArgs{
		Name: pulumi.Sprintf("/aws/lambda/%s", lambdaFunction.Name),
	})
	if err != nil {
		return err
	}

	return nil
}
