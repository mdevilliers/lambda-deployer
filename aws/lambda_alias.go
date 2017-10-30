package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	deployer "github.com/mdevilliers/lambda-deployer"
	"github.com/pkg/errors"
)

func CreateOrUpdateAlias(svc *lambda.Lambda, conf *lambda.FunctionConfiguration, meta deployer.FunctionMetadata) error {
	exists, err := aliasExists(svc, meta.FunctionName, meta.Alias)

	if err != nil {
		return err
	}

	if exists {
		return updateAlias(svc, meta.FunctionName, meta.Alias, *conf.Version)
	}

	return newAlias(svc, meta.FunctionName, meta.Alias, *conf.Version)
}

func updateAlias(svc *lambda.Lambda, functionName, aliasName, functionVersion string) error {

	req := &lambda.UpdateAliasInput{
		FunctionName:    aws.String(functionName),
		Name:            aws.String(aliasName),
		FunctionVersion: aws.String(functionVersion),
	}

	resp, err := svc.UpdateAlias(req)

	log.Println("UpdateAlias : ", resp, err)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil

}

func newAlias(svc *lambda.Lambda, functionName, aliasName, functionVersion string) error {

	req := &lambda.CreateAliasInput{
		FunctionName:    aws.String(functionName),
		Name:            aws.String(aliasName),
		FunctionVersion: aws.String(functionVersion),
	}

	resp, err := svc.CreateAlias(req)

	log.Println("CreateAlias : ", resp, err)

	if err != nil {
		return errors.WithStack(err)
	}

	return nil

}

func aliasExists(svc *lambda.Lambda, functionName, aliasName string) (bool, error) {

	req := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(aliasName),
	}

	resp, err := svc.GetAlias(req)

	log.Println("GetAlias : ", resp, err)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeResourceNotFoundException:
				return false, nil
			}
			return false, err
		}
	}

	return true, nil
}
