package aws

import (
	"fmt"
	"log"
	"sort"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	deployer "github.com/mdevilliers/lambda-deployer"
	"github.com/pkg/errors"
)

func CreateOrUpdateFunction(svc *lambda.Lambda, bucket, key, role string, meta deployer.FunctionMetadata) (*lambda.FunctionConfiguration, error) {

	exists, err := functionExists(svc, meta.FunctionName)

	if err != nil {
		return nil, err
	}

	if exists {

		return updateLambdaFunction(svc, bucket, key, meta)

	}

	return newLambdaFunction(svc, bucket, key, role, meta)

}

func functionExists(svc *lambda.Lambda, name string) (bool, error) {

	req := &lambda.GetFunctionInput{
		FunctionName: aws.String(name),
	}

	_, err := svc.GetFunction(req)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case lambda.ErrCodeResourceNotFoundException:
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func updateLambdaFunction(svc *lambda.Lambda, s3Bucket, s3Key string, metadata deployer.FunctionMetadata) (*lambda.FunctionConfiguration, error) {

	req := &lambda.UpdateFunctionCodeInput{
		FunctionName: aws.String(metadata.FunctionName),
		Publish:      aws.Bool(true),
		S3Bucket:     aws.String(s3Bucket),
		S3Key:        aws.String(s3Key),
	}

	resp, err := svc.UpdateFunctionCode(req)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}

func newLambdaFunction(svc *lambda.Lambda, s3Bucket, s3Key, role string, metadata deployer.FunctionMetadata) (*lambda.FunctionConfiguration, error) {

	req := &lambda.CreateFunctionInput{
		Code: &lambda.FunctionCode{
			S3Bucket: aws.String(s3Bucket),
			S3Key:    aws.String(s3Key),
		},
		Description:  aws.String(metadata.Description),
		FunctionName: aws.String(metadata.FunctionName),
		Handler:      aws.String(metadata.Handler),
		MemorySize:   aws.Int64(metadata.MemorySize),
		Publish:      aws.Bool(true),
		Role:         aws.String(role),
		Runtime:      aws.String(metadata.Runtime),
		Timeout:      aws.Int64(metadata.Timeout),
		Environment: &lambda.Environment{
			Variables: map[string]*string{},
		},
	}

	for k, v := range metadata.EnvVars {
		req.Environment.Variables[k] = aws.String(fmt.Sprintf("%v", v))
	}

	resp, err := svc.CreateFunction(req)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return resp, nil
}

func ReduceUnAliasedVersions(svc *lambda.Lambda, maxVersions int, metadata deployer.FunctionMetadata) error {

	// get all aliased functions
	allAliasesReq := &lambda.ListAliasesInput{
		FunctionName: aws.String(metadata.FunctionName),
	}
	allAliasResp, err := svc.ListAliases(allAliasesReq)

	if err != nil {
		errors.WithStack(err)
	}

	// get all versions for a function
	versionReq := &lambda.ListVersionsByFunctionInput{
		FunctionName: aws.String(metadata.FunctionName),
	}

	versionResp, err := svc.ListVersionsByFunction(versionReq)

	if err != nil {
		errors.WithStack(err)
	}

	// if there are less (or equal) versions than the max versions
	if len(versionResp.Versions) <= maxVersions {
		return nil
	}

	// create an array to hold all versions without an active alias
	versionsUnAliased := []*lambda.FunctionConfiguration{}

	// use the code hash to build up a list of versions without aliases
	for _, version := range versionResp.Versions {

		drop := false

		// $LATEST is a special pointer to the latest function
		// helpfully it isn't returned in the list of aliases
		// so we need a special case here
		if *(version.Version) == "$LATEST" {
			drop = true
		} else {
			// we need to loop though our list of aliases checking if that version
			// hasn't been assigned an alias
			for _, aliasedFunction := range allAliasResp.Aliases {

				if *(aliasedFunction.FunctionVersion) == *(version.Version) {
					drop = true
				}
			}
		}

		if !drop {
			versionsUnAliased = append(versionsUnAliased, version)
		}

	}

	// if the unaliased versions number less or equal to maxVersions to retain
	if len(versionsUnAliased) <= maxVersions {
		return nil
	}

	// order by versions
	sort.Sort(byVersion(versionsUnAliased))

	// delete all versions - the last n (maxVersions)
	toDelete := versionsUnAliased[0 : len(versionsUnAliased)-maxVersions]

	for _, version := range toDelete {

		log.Println("deleting unaliased function : ", *(version.Version))
		deleteRequest := &lambda.DeleteFunctionInput{
			FunctionName: version.FunctionName,
			Qualifier:    version.Version,
		}
		_, err := svc.DeleteFunction(deleteRequest)

		if err != nil {
			return errors.WithStack(err)
		}

	}

	return nil
}

// byVersion implements sort.Interface for []*lambda.FunctionConfiguration based on
// the Version.
type byVersion []*lambda.FunctionConfiguration

func (a byVersion) Len() int      { return len(a) }
func (a byVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byVersion) Less(i, j int) bool {

	iVersion, err := strconv.Atoi(*(a[i].Version))

	if err != nil {
		panic("version not a number")
	}

	jVersion, _ := strconv.Atoi(*(a[j].Version))

	if err != nil {
		panic("version not a number")
	}

	return iVersion < jVersion
}
