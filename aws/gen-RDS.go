package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (aws *AWS) RDSDescribeDBInstances() (error, RDSDescribeDBInstancesResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws rds describe-db-instances --region %s ", aws.Region))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp RDSDescribeDBInstancesResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) RDSDeleteDBInstance(dbInstanceIdentifier string) (error, RDSCreateDBInstanceResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws rds delete-db-instance --region %s --db-instance-identifier %s --skip-final-snapshot", aws.Region, dbInstanceIdentifier))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp RDSCreateDBInstanceResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) RDSCreateDBInstance(instanceName, dbIdentifier, masterUsername, masterPassword, availabilityZone, subnetGroupName, port, engineVersion, dbInstanceClass, storageSize string) (error, RDSCreateDBInstanceResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws rds create-db-instance --region %s --db-name %s --db-instance-identifier %s --master-username %s --master-user-password %s --availability-zone %s --db-subnet-group-name %s --port %s --engine-version %s --no-publicly-accessible --engine mysql --db-instance-class %s --allocated-storage %s", aws.Region, instanceName, dbIdentifier, masterUsername, masterPassword, availabilityZone, subnetGroupName, port, engineVersion, dbInstanceClass, storageSize))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp RDSCreateDBInstanceResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) RDSDescribeSubnetGroups() (error, RDSDescribeSubnetGroupsResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws rds describe-db-subnet-groups --region %s ", aws.Region))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp RDSDescribeSubnetGroupsResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) RDSCreateSubnetGroup(subnetGroupName, subnetGroupDescription, subnetIds string) (error, RDSCreateDBSubnetGroupResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws rds create-db-subnet-group --region %s --db-subnet-group-name %s --db-subnet-group-description %s --subnet-ids %s", aws.Region, subnetGroupName, subnetGroupDescription, subnetIds))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp RDSCreateDBSubnetGroupResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) RDSDeleteSubnetGroup(subnetGroupName string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws rds delete-db-subnet-group --region %s --db-subnet-group-name %s", aws.Region, subnetGroupName))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}
