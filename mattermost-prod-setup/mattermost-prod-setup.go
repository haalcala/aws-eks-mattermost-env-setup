package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"../aws"
)

var REGION = "us-east-2"

var CLUSTER_KEY = "mattermost-prod"
var DB_SUBNET_GROUP_NAME = CLUSTER_KEY + "-subg"
var DB_INSTANCE_ID = CLUSTER_KEY + "-db"
var DB_INSTANCE_DB_NAME = strings.ReplaceAll(CLUSTER_KEY, "-", "")
var SUBNET_NAME_PREFIX = CLUSTER_KEY + "-sub-"
var NAT_NAME_PREFIX = CLUSTER_KEY + "-nat-"
var IGW_NAME = CLUSTER_KEY + "-igw"
var ROUTE_TABLE_PREFIX = CLUSTER_KEY + "-rt-"
var ROUTE_PREFIX = CLUSTER_KEY + "-rt-"
var VPC_NAME = CLUSTER_KEY + "-vpc"
var STACK_NAME = fmt.Sprintf("eksctl-%s-cluster", CLUSTER_KEY)

var VPC_ID string

func getVpc(vpcs []aws.AWSVPCType) *aws.AWSVPCType {
	var vpc *aws.AWSVPCType = nil

	if len(vpcs) > 0 {
		for _, _vpc := range vpcs {
			for _, _tag := range _vpc.Tags {
				if _tag.Key == "Name" && _tag.Value == VPC_NAME {
					vpc = &_vpc
				}
			}
		}
	}

	return vpc
}

func createVPC(awsm *aws.AWS) (error, *aws.AWSVPCType) {
	err, vpc := awsm.CreateVPC("main", "10.0.0.0/16")

	if err != nil {
		fmt.Println("err:", err)
	}

	createTag(awsm, vpc.VpcId, "Name", fmt.Sprintf("%s-vpc", CLUSTER_KEY), "kubernetes.io/cluster/"+CLUSTER_KEY, "shared", "alpha.eksctl.io/cluster-name", CLUSTER_KEY)

	if err != nil {
		fmt.Println("err:", err)
	}

	for vpc != nil && vpc.State != "available" {

		fmt.Println("vpc:", vpc)

		time.Sleep(5 * time.Second)

		err, vpcs := awsm.GetVPCs()
		vpc = getVpc(vpcs)

		if err != nil {
			return err, nil
		}
	}

	return err, vpc
}

func getSubnets(awsm *aws.AWS, vpc aws.AWSVPCType) (error, []aws.AWSSubnetType) {
	err, subnets := awsm.GetSubnetsByVPCId(vpc.VpcId)

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("subnets:", subnets)

	return err, subnets
}

func createSubnet(awsm *aws.AWS, vpc aws.AWSVPCType, zones []string) (error, []aws.AWSSubnetType) {
	categories := []string{"public", "private", "internal"}

	for zonei, zone := range zones {
		fmt.Println("zonei:", zonei)

		categories_i := 0
		for _, subnet := range Subnets[(zonei * len(categories)) : (zonei+1)*len(categories)] {

			fmt.Println("zone:", zone, "subnet:", subnet)

			err, subnet := awsm.CreateSubnet(vpc.VpcId, subnet.CIDR+"/21", zone)

			if err != nil {
				fmt.Println("err:", err)
				os.Exit(1)
			}

			tags := []string{"Name", fmt.Sprintf("%s-sub-%s-%d", CLUSTER_KEY, categories[categories_i], zonei+1), "kubernetes.io/cluster/" + CLUSTER_KEY, "shared", "alpha.eksctl.io/cluster-name", CLUSTER_KEY}

			if categories[categories_i] == "public" {
				// err, createDefaultSubnetResponse := awsm.CreateDefaultSubnet(zone)

				// if err != nil {
				// 	fmt.Println("err:", err)
				// 	os.Exit(1)
				// }

				// fmt.Println("createDefaultSubnetResponse:", createDefaultSubnetResponse)

				tags = append(tags, "kubernetes.io/role/elb", "1")
			} else if categories[categories_i] == "private" {
				tags = append(tags, "kubernetes.io/role/internal-elb", "1")
			}

			createTag(awsm, subnet.SubnetId, tags...)

			categories_i = categories_i + 1
		}
	}

	err, subnets := getSubnets(awsm, vpc)

	return err, subnets
}

func getInternetGateway(awsm *aws.AWS) (error, *aws.AWSInternetGatewayType) {
	err, describeInternetGatewaysResponse := awsm.EC2DescribeInternetGateways()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("describeInternetGatewaysResponse:", describeInternetGatewaysResponse)

	var igw *aws.AWSInternetGatewayType = nil

	for _, _igw := range describeInternetGatewaysResponse.InternetGateways {
		for _, _tag := range _igw.Tags {
			if _tag.Key == "Name" && _tag.Value == IGW_NAME {
				igw = &_igw
				break
			}
		}
	}

	return err, igw
}

func getNatGateways(awsm *aws.AWS) (error, []aws.AWSNatGatewayType) {
	err, ngws := awsm.GetNatGateways()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("ngws:", ngws)

	var ret = []aws.AWSNatGatewayType{}

	for _, _ngw := range ngws {
		if _ngw.State != "deleted" && _ngw.State != "deleting" {
			for _, _tag := range _ngw.Tags {
				if _tag.Key == "Name" && strings.HasPrefix(_tag.Value, NAT_NAME_PREFIX) {
					ret = append(ret, _ngw)
				}
			}
		}
	}

	return err, ret
}

func createInternetGateway(awsm *aws.AWS, vpc aws.AWSVPCType) (error, *aws.AWSInternetGatewayType) {
	err, igw := awsm.CreateInternetGateway()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	createTag(awsm, igw.InternetGatewayId, "Name", fmt.Sprintf("%s-igw", CLUSTER_KEY), "alpha.eksctl.io/cluster-name", CLUSTER_KEY)

	_ = awsm.AttachInternetGatewayToVPC(vpc.VpcId, igw.InternetGatewayId)

	return err, igw
}

func createNatGateway(awsm *aws.AWS, subnets []aws.AWSSubnetType, addresses []aws.AWSAddressType) (error, []aws.AWSNatGatewayType) {
	fmt.Println("---===>>> createNatGateway:")

	err, ngws := getNatGateways(awsm)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	var available_addresses []aws.AWSAddressType

	for _, address := range addresses {
		if address.AssociationId == nil {
			available_addresses = append(available_addresses, address)
		}
	}

	fmt.Println("available_addresses:", available_addresses)

	for _, subnet := range subnets {
		for _, tag := range subnet.Tags {
			if tag.Key == "Name" && strings.HasPrefix(tag.Value, fmt.Sprintf("%spublic-", SUBNET_NAME_PREFIX)) {
				fmt.Println("Found public subnet:", subnet)

				found := false

				for _, ngw := range ngws {
					if ngw.SubnetId == subnet.SubnetId {
						found = true
						break
					}
				}

				if !found {
					selected_address := available_addresses[0]
					available_addresses = available_addresses[1:]

					fmt.Println("selected_address:", selected_address)

					err, createNatResponse := awsm.CreateNatGateway(*selected_address.AllocationId, subnet.SubnetId)

					if err != nil {
						fmt.Println("error:", err)
						os.Exit(1)
					}

					fmt.Println("createNatResponse:", createNatResponse)

					createTag(awsm, createNatResponse.NatGatewayId, "Name", fmt.Sprintf("%s%s", NAT_NAME_PREFIX, strings.Split(tag.Value, "public")[1]), "alpha.eksctl.io/cluster-name", CLUSTER_KEY)
				}
			}
		}
	}

	err, ngws = getNatGateways(awsm)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	return err, ngws
}

func getAddresses(awsm *aws.AWS) (error, []aws.AWSAddressType) {
	err, addresses := awsm.GetAddresses()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	return err, addresses.Addresses
}

func getRoutes(awsm *aws.AWS) (error, []aws.AWSRouteTableType) {
	err, routes := awsm.GetRouteTables()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	return err, routes.RouteTables
}

func createTag(awsm *aws.AWS, resourceId string, KeyValuePair ...string) {
	tags := []string{}

	for i := range make([]int, len(KeyValuePair)/2) {
		tags = append(tags, fmt.Sprintf("Key=%s,Value=%s", KeyValuePair[i*2], KeyValuePair[i*2+1]))
	}

	tags = append(tags, fmt.Sprintf("Key=%s,Value=%s", "CreatedBy", "mattermost-prod-setup.go"))

	_ = awsm.CreateTags(resourceId, strings.Join(tags, " "))
}

func analyseRoutes(awsm *aws.AWS, vpc aws.AWSVPCType, subnets []aws.AWSSubnetType) {
	fmt.Println("***************************************************************************")
	fmt.Println("***************************************************************************")
	fmt.Println("Analysing routes ...")

	err, subnets := getSubnets(awsm, vpc)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("subnets:", subnets)

	err, igw := getInternetGateway(awsm)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("igw:", igw)

	err, ngws := getNatGateways(awsm)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("ngws:", ngws)

	err, routes := getRoutes(awsm)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("routes:", routes)

	categories := []string{"public", "private", "internal"}

	for _, subnet := range subnets {
		fmt.Println("subnet:", subnet)
		for _, tag := range subnet.Tags {
			if tag.Key == "Name" && strings.HasPrefix(tag.Value, SUBNET_NAME_PREFIX) {

				// find whether the subnet is private|public|internal (category)
				for _, category := range categories {
					subnet_key := strings.Split(tag.Value, SUBNET_NAME_PREFIX)[1]

					fmt.Println("subnet_key:", subnet_key)

					if strings.HasPrefix(subnet_key, category) {
						found := false
						for _, route := range routes {
							for _, route_association := range route.Associations {
								if route_association.SubnetId == subnet.SubnetId {
									found = true
									break
								}
							}
						}

						if !found {
							err, route := awsm.CreateRouteTable(vpc)

							if err != nil {
								fmt.Println("error:", err)
								os.Exit(1)
							}

							fmt.Println("route:", route)

							createTag(awsm, route.RouteTable.RouteTableId, "Name", fmt.Sprintf("%s-rt%s", CLUSTER_KEY, strings.Split(tag.Value, fmt.Sprintf("%s-sub", CLUSTER_KEY))[1]), "alpha.eksctl.io/cluster-name", CLUSTER_KEY)

							err, associateTableResponse := awsm.AssociateRouteTable(route.RouteTable.RouteTableId, subnet.SubnetId)

							if err != nil {
								fmt.Println("error:", err)
								os.Exit(1)
							}

							fmt.Println("associateTableResponse:", associateTableResponse)

							if category == "public" {
								err, createRouteResult := awsm.CreateRouteWithInternetGateway(route.RouteTable.RouteTableId, "0.0.0.0/0", igw.InternetGatewayId)

								if err != nil {
									fmt.Println("error:", err)
									os.Exit(1)
								}

								fmt.Println("createRouteResult:", createRouteResult)
							} else {
								found := false

								fmt.Println("Trying to look for Nat gateway in the same availability zone", subnet.AvailabilityZone)
								for _, ngw := range ngws {
									fmt.Println("---- ngw:", ngw)

									// find this nat gateway's subnet
									for _, _subnet := range subnets {
										fmt.Println("---- _subnet:", _subnet)

										if ngw.SubnetId == _subnet.SubnetId {

											if subnet.AvailabilityZone == _subnet.AvailabilityZone {
												found = true

												fmt.Println("Found nat gateway for the same availability zone!!!")
												err, createRouteResult := awsm.CreateRouteWithNatGateway(route.RouteTable.RouteTableId, "0.0.0.0/0", ngw.NatGatewayId)

												if err != nil {
													fmt.Println("error:", err)
													os.Exit(1)
												}

												fmt.Println("createRouteResult:", createRouteResult)

												break
											}
										}

										if found {
											break
										}
									}

									if found {
										break
									}
								}
							}
						}
					}
				}
			}
		}
	}
}

func deployDatabase(awsm *aws.AWS, subnets []aws.AWSSubnetType) {
	err, subnetGroups := awsm.RDSDescribeSubnetGroups()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("subnetGroups:", subnetGroups)

	var subnetGroup *aws.AWSRDSSubnetGroupType = nil

	for _, _subnetGroup := range subnetGroups.DBSubnetGroups {
		for _, tag := range _subnetGroup.Tags {
			if tag.Key == "Name" && strings.HasPrefix(tag.Value, fmt.Sprintf("%s-", CLUSTER_KEY)) {
				subnetGroup = &_subnetGroup
			}
		}
	}

	fmt.Println("subnetGroup:", subnetGroup)

	subnetIds := []string{}
	var firstPublicSubnet *aws.AWSSubnetType = nil

	if subnetGroup == nil {
		for _, subnet := range subnets {
			fmt.Println("subnet:", subnet)
			for _, tag := range subnet.Tags {
				if tag.Key == "Name" && strings.HasPrefix(tag.Value, fmt.Sprintf("%sinternal-", SUBNET_NAME_PREFIX)) {
					subnetIds = append(subnetIds, subnet.SubnetId)
				}
				if firstPublicSubnet == nil && tag.Key == "Name" && strings.HasPrefix(tag.Value, fmt.Sprintf("%spublic-", SUBNET_NAME_PREFIX)) {
					firstPublicSubnet = &subnet
				}
			}
		}

		err, createSubnetGroupResponse := awsm.RDSCreateSubnetGroup(DB_SUBNET_GROUP_NAME, DB_SUBNET_GROUP_NAME, strings.Join(subnetIds, " "))

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("createSubnetGroupResponse:", createSubnetGroupResponse)

		subnetGroup = &createSubnetGroupResponse.DBSubnetGroup
	}

	err, describeDbInstancesResponse := awsm.RDSDescribeDBInstances()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("describeDbInstancesResponse:", describeDbInstancesResponse)

	var dbInstance *aws.AWSDBInstanceType = nil

	for _, _dbInstance := range describeDbInstancesResponse.DBInstances {
		if _dbInstance.DBInstanceIdentifier == DB_INSTANCE_ID {
			dbInstance = &_dbInstance
		}
	}

	fmt.Println("dbInstance:", dbInstance)

	if dbInstance == nil {
		err, createDBResponse := awsm.RDSCreateDBInstance(DB_INSTANCE_DB_NAME, DB_INSTANCE_ID, "admin", "vcube2192", firstPublicSubnet.AvailabilityZone, DB_SUBNET_GROUP_NAME, "3306", "5.7.26", "db.t2.micro", "20")

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("createDBResponse:", createDBResponse)

		dbInstance = &createDBResponse.DBInstance

		fmt.Println("dbInstance:", dbInstance)
	}

	for {
		if dbInstance.DBInstanceStatus == "available" {
			break
		}

		time.Sleep(5 * time.Second)

		err, describeDbInstancesResponse := awsm.RDSDescribeDBInstances()

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("describeDbInstancesResponse:", describeDbInstancesResponse)

		for _, _dbInstance := range describeDbInstancesResponse.DBInstances {
			if _dbInstance.DBInstanceIdentifier == DB_INSTANCE_ID {
				dbInstance = &_dbInstance
			}
		}
	}
}

func GetRds(awsm *aws.AWS) (error, *aws.RDSPayload) {
	err, getRdsResponse := awsm.GetRDS()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("getRdsResponse:", getRdsResponse)

	log.Println("get_rds_response.DBInstances:", reflect.TypeOf(getRdsResponse), reflect.TypeOf(getRdsResponse.DBInstances), getRdsResponse.DBInstances)

	requestBody, err := json.Marshal(getRdsResponse)

	log.Println("requestBody:", string(requestBody))

	var rds_instance *aws.RDSPayload = nil

	fmt.Println("len(getRdsResponse.DBInstances):", len(getRdsResponse.DBInstances))

	if len(getRdsResponse.DBInstances) > 0 {
		for _, _rds_instance := range getRdsResponse.DBInstances {
			fmt.Println("1111 rds_instance:", _rds_instance)

			if _rds_instance.DBInstanceIdentifier == CLUSTER_KEY {
				rds_instance = &_rds_instance
			}
		}
	}

	return err, rds_instance
}

func createStack(awsm *aws.AWS, subnets []aws.AWSSubnetType) {

	err, stack := awsm.GetStackWithTag("alpha.eksctl.io/cluster-name", CLUSTER_KEY)

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("stack:", stack)

	for stack != nil && stack.StackStatus == "DELETE_IN_PROGRESS" {
		fmt.Println("Previous stack detected.")

		time.Sleep(10 * time.Second)

		err, stack = awsm.GetStackWithTag("alpha.eksctl.io/cluster-name", CLUSTER_KEY)

		if err != nil {
			fmt.Println("err:", err)
		}
	}

	var publicSubnets, privateSubnets = []string{}, []string{}

	for _, subnet := range subnets {
		for _, tag := range subnet.Tags {
			if tag.Key == "Name" {
				if strings.HasPrefix(tag.Value, fmt.Sprintf("%spublic-", SUBNET_NAME_PREFIX)) {
					publicSubnets = append(publicSubnets, subnet.SubnetId)
				} else if strings.HasPrefix(tag.Value, fmt.Sprintf("%sinternal-", SUBNET_NAME_PREFIX)) {
					privateSubnets = append(privateSubnets, subnet.SubnetId)
				}
			}
		}
	}

	if stack == nil {
		err := awsm.EKSCreateFargateCluster(CLUSTER_KEY, strings.Join(privateSubnets, ","), strings.Join(publicSubnets, ","))

		if err != nil {
			fmt.Println("err:", err)
		}

		for {
			err, stack = awsm.GetStackWithTag("alpha.eksctl.io/cluster-name", CLUSTER_KEY)

			if err != nil {
				fmt.Println("err:", err)
				break
			}

			if stack != nil && stack.StackStatus != "CREATE_COMPLETE" {
				time.Sleep(10 * time.Second)

				continue
			}

			break
		}
	}

	err, fargateProfiles := awsm.EKSListFargateProfiles(CLUSTER_KEY)

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("fargateProfiles:", fargateProfiles.FargateProfileNames)

	// fargateProfileName := ""

	// for _, _fargateProfile := range fargateProfiles.FargateProfileNames {
	// 	err, fargateProfile := awsm.EKSDescribeFargateProfile(CLUSTER_KEY, _fargateProfile)

	// 	if err != nil {
	// 		fmt.Println("err:", err)
	// 	}

	// 	fmt.Println("fargateProfile:", fargateProfile)

	// 	if fargateProfile.FargateProfile.FargateProfileName == "fp-default" && fargateProfile.FargateProfile.Status == "ACTIVE" {
	// 		err, deleteFPProfileResponse := awsm.EKSDeleteFargateProfile(CLUSTER_KEY, fargateProfile.FargateProfile.FargateProfileName)

	// 		if err != nil {
	// 			fmt.Println("err:", err)
	// 		}

	// 		fmt.Println("deleteFPProfileResponse:", deleteFPProfileResponse)
	// 	}
	// }

	// err, eksCluster := awsm.EKSDescribeCluster(CLUSTER_KEY)

	// if err != nil {
	// 	fmt.Println("err:", err)
	// }

	// fmt.Println("eksCluster:", eksCluster)

	// err, fargateExecutionRole := getEksPodExecutionRole(awsm)

	// if err != nil {
	// 	fmt.Println("err:", err)
	// }

	// fmt.Println("fargateExecutionRole:", fargateExecutionRole)

	// if fargateExecutionRole != nil {
	// 	err, createFargateProfileResponse := awsm.EKSCreateFargateProfile(CLUSTER_KEY, "demo-default", fargateExecutionRole.Arn, "namespace=default namespace=kube-system", strings.Join(privateSubnets, " "), "harold=alcala")

	// 	if err != nil {
	// 		fmt.Println("err:", err)
	// 	}

	// 	fmt.Println("createFargateProfileResponse:", createFargateProfileResponse)
	// }
}

func getEksPodExecutionRole(awsm *aws.AWS) (error, *aws.AWSIAMRoleType) {
	err, listRolesResponse := awsm.IAMListRoles()

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("listRolesResponse:", listRolesResponse)

	var role *aws.AWSIAMRoleType = nil

	for _, _role := range listRolesResponse.Roles {
		fmt.Println("_role:", _role)
		fmt.Println("_role.RoleName:", _role.RoleName)
		fmt.Println("fmt.Sprintf(\"eksctl-%%s-clu-FargatePodExecutionRole-\", CLUSTER_KEY):", fmt.Sprintf("eksctl-%s-clu-FargatePodExecutionRole-", CLUSTER_KEY))
		if strings.HasPrefix(_role.RoleName, fmt.Sprintf("eksctl-%s-clu-FargatePodExecutionRole-", CLUSTER_KEY)) {
			role = &_role
			break
		}
	}

	fmt.Println("role:", role)

	return err, role
}

func doCleanup(awsm *aws.AWS) {
	err, fargateProfiles := awsm.EKSListFargateProfiles(CLUSTER_KEY)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("fargateProfiles:", fargateProfiles)

	for _, _fargateProfile := range fargateProfiles.FargateProfileNames {
		err, fargateProfile := awsm.EKSDescribeFargateProfile(CLUSTER_KEY, _fargateProfile)

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("fargateProfile:", fargateProfile)

		if fargateProfile.FargateProfile.Status == "ACTIVE" {
			err, deleteFargateProfileResponse := awsm.EKSDeleteFargateProfile(CLUSTER_KEY, fargateProfile.FargateProfile.FargateProfileName)

			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}

			fmt.Println("deleteFargateProfileResponse:", deleteFargateProfileResponse)
		}

		for {
			time.Sleep(10 * time.Second)

			err, fargateProfile = awsm.EKSDescribeFargateProfile(CLUSTER_KEY, _fargateProfile)

			if fargateProfile.FargateProfile.Status == "DELETING" {
				continue
			}

			break
		}
	}

	err, clusters := awsm.EKSListClusters()

	cluster := ""

	for _, _cluster := range clusters.Clusters {
		if _cluster == CLUSTER_KEY {
			cluster = _cluster
		}
	}

	if cluster != "" {
		err, describeClusterResponse := awsm.EKSDescribeCluster(CLUSTER_KEY)

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("describeClusterResponse:", describeClusterResponse)

		if describeClusterResponse.Cluster.Status == "ACTIVE" {
			err, deleteClusterResponse := awsm.EKSDeleteCluster(cluster)

			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}

			fmt.Println("deleteClusterResponse:", deleteClusterResponse)
		}

		for {

			err, describeClusterResponse = awsm.EKSDescribeCluster(CLUSTER_KEY)

			fmt.Println("describeClusterResponse.Cluster.Status:", describeClusterResponse.Cluster.Status)

			if describeClusterResponse.Cluster.Status == "DELETING" {
				time.Sleep(10 * time.Second)

				continue
			}

			break
		}
	}

	err, rds_instances := awsm.RDSDescribeDBInstances()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("rds_instances:", rds_instances)

	for _, _rds := range rds_instances.DBInstances {
		if _rds.DBInstanceIdentifier == DB_INSTANCE_ID {
			err, deleteRdsInstanceResponse := awsm.RDSDeleteDBInstance(DB_INSTANCE_ID)

			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}

			fmt.Println("deleteRdsInstanceResponse:", deleteRdsInstanceResponse)
		}
	}

	for {
		err, rds_instances := awsm.RDSDescribeDBInstances()

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("rds_instances:", rds_instances)

		var dbInstance *aws.AWSDBInstanceType = nil

		for _, _rds := range rds_instances.DBInstances {
			if _rds.DBInstanceIdentifier == DB_INSTANCE_ID {
				dbInstance = &_rds

			}
		}

		if dbInstance != nil {
			if strings.HasPrefix(dbInstance.DBInstanceStatus, "delet") {
				fmt.Println("Waiting for the RDS instance to be fully deleted...")
				time.Sleep(10 * time.Second)
				continue
			}
		}

		break
	}

	err, rdsSubnetGroups := awsm.RDSDescribeSubnetGroups()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("rdsSubnetGroups:", rdsSubnetGroups)

	var subnetGroup *aws.AWSRDSSubnetGroupType = nil

	for _, _subnetGroup := range rdsSubnetGroups.DBSubnetGroups {
		if _subnetGroup.DBSubnetGroupName == DB_SUBNET_GROUP_NAME {
			subnetGroup = &_subnetGroup
		}
	}

	if subnetGroup != nil {
		err = awsm.RDSDeleteSubnetGroup(subnetGroup.DBSubnetGroupName)

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

	}

	err, vpcs := awsm.GetVPCs()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	vpc := getVpc(vpcs)

	err, routeTables := getRoutes(awsm)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("routeTables:", routeTables)

	for _, routeTable := range routeTables {
		for _, tag := range routeTable.Tags {
			if tag.Key == "Name" && strings.HasPrefix(tag.Value, fmt.Sprintf("%s-", CLUSTER_KEY)) {
				if routeTable.Associations != nil && len(routeTable.Associations) > 0 {
					for _, association := range routeTable.Associations {
						if !association.Main {
							err = awsm.DisassociateRouteTable(association.RouteTableAssociationId)

							if err != nil {
								fmt.Println("error:", err)
								os.Exit(1)
							}
						}
					}
				}

				err = awsm.DeleteRouteTable(routeTable.RouteTableId)

				if err != nil {
					fmt.Println("error:", err)
					os.Exit(1)
				}
			}
		}
	}

	err, natGateways := getNatGateways(awsm)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("natGateways:", natGateways)

	for _, natGateway := range natGateways {
		err = awsm.DeleteNatGateway(natGateway.NatGatewayId)

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}

	deleteInternetGateway(awsm)

	if vpc != nil {
		err, subnets := awsm.GetSubnetsByVPCId(vpc.VpcId)

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}

		fmt.Println("subnets:", subnets)

		for _, subnet := range subnets {
			for _, tag := range subnet.Tags {
				if tag.Key == "Name" && strings.HasPrefix(tag.Value, fmt.Sprintf("%s-", CLUSTER_KEY)) {
					err = awsm.DeleteSubnet(subnet.SubnetId)

					if err != nil {
						fmt.Println("error:", err)
						os.Exit(1)
					}
				}
			}
		}

		deleteVPC(awsm)
	}

	deleteStack(awsm)
}

func deleteInternetGateway(awsm *aws.AWS) {
	err, igw := getInternetGateway(awsm)

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	fmt.Println("igw:", igw)

	if igw != nil {
		for _, attachment := range igw.Attachments {
			_ = awsm.EC2DetachInternetGateway(igw.InternetGatewayId, attachment.VpcId)
		}

		err = awsm.EC2DeleteInternetGateway(igw.InternetGatewayId)

		if err != nil {
			fmt.Println("error:", err)
			os.Exit(1)
		}
	}
}

func deleteStack(awsm *aws.AWS) {
	err, describeStacksResponse := awsm.CFDescribeStacks()

	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	for _, stack := range describeStacksResponse.Stacks {
		if stack.StackName == STACK_NAME {
			err = awsm.CFDeleteStack(STACK_NAME)

			if err != nil {
				fmt.Println("error:", err)
				os.Exit(1)
			}
		}
	}
}

func deleteVPC(awsm *aws.AWS) {
	err, vpcs := awsm.GetVPCs()

	if err != nil {
		fmt.Println("err:", err)
	}

	vpc := getVpc(vpcs)

	err = awsm.EC2DeleteVPC(vpc.VpcId)
}

// CIDR 10.0.0.0/21 networks
var Subnets = []aws.Subnet{
	aws.Subnet{"10.0.0.0", "10.0.0.1", "10.0.7.254", "10.0.7.255"},
	aws.Subnet{"10.0.8.0", "10.0.8.1", "10.0.15.254", "10.0.15.255"},
	aws.Subnet{"10.0.16.0", "10.0.16.1", "10.0.23.254", "10.0.23.255"},
	aws.Subnet{"10.0.24.0", "10.0.24.1", "10.0.31.254", "10.0.31.255"},
	aws.Subnet{"10.0.32.0", "10.0.32.1", "10.0.39.254", "10.0.39.255"},
	aws.Subnet{"10.0.40.0", "10.0.40.1", "10.0.47.254", "10.0.47.255"},
	aws.Subnet{"10.0.48.0", "10.0.48.1", "10.0.55.254", "10.0.55.255"},
	aws.Subnet{"10.0.56.0", "10.0.56.1", "10.0.63.254", "10.0.63.255"},
	aws.Subnet{"10.0.64.0", "10.0.64.1", "10.0.71.254", "10.0.71.255"},
	aws.Subnet{"10.0.72.0", "10.0.72.1", "10.0.79.254", "10.0.79.255"},
	aws.Subnet{"10.0.80.0", "10.0.80.1", "10.0.87.254", "10.0.87.255"},
	aws.Subnet{"10.0.88.0", "10.0.88.1", "10.0.95.254", "10.0.95.255"},
	aws.Subnet{"10.0.96.0", "10.0.96.1", "10.0.103.254", "10.0.103.255"},
	aws.Subnet{"10.0.104.0", "10.0.104.1", "10.0.111.254", "10.0.111.255"},
	aws.Subnet{"10.0.112.0", "10.0.112.1", "10.0.119.254", "10.0.119.255"},
	aws.Subnet{"10.0.120.0", "10.0.120.1", "10.0.127.254", "10.0.127.255"},
	aws.Subnet{"10.0.128.0", "10.0.128.1", "10.0.135.254", "10.0.135.255"},
	aws.Subnet{"10.0.136.0", "10.0.136.1", "10.0.143.254", "10.0.143.255"},
	aws.Subnet{"10.0.144.0", "10.0.144.1", "10.0.151.254", "10.0.151.255"},
	aws.Subnet{"10.0.152.0", "10.0.152.1", "10.0.159.254", "10.0.159.255"},
	aws.Subnet{"10.0.160.0", "10.0.160.1", "10.0.167.254", "10.0.167.255"},
	aws.Subnet{"10.0.168.0", "10.0.168.1", "10.0.175.254", "10.0.175.255"},
	aws.Subnet{"10.0.176.0", "10.0.176.1", "10.0.183.254", "10.0.183.255"},
	aws.Subnet{"10.0.184.0", "10.0.184.1", "10.0.191.254", "10.0.191.255"},
	aws.Subnet{"10.0.192.0", "10.0.192.1", "10.0.199.254", "10.0.199.255"},
	aws.Subnet{"10.0.200.0", "10.0.200.1", "10.0.207.254", "10.0.207.255"},
	aws.Subnet{"10.0.208.0", "10.0.208.1", "10.0.215.254", "10.0.215.255"},
	aws.Subnet{"10.0.216.0", "10.0.216.1", "10.0.223.254", "10.0.223.255"},
	aws.Subnet{"10.0.224.0", "10.0.224.1", "10.0.231.254", "10.0.231.255"},
	aws.Subnet{"10.0.232.0", "10.0.232.1", "10.0.239.254", "10.0.239.255"},
	aws.Subnet{"10.0.240.0", "10.0.240.1", "10.0.247.254", "10.0.247.255"},
	aws.Subnet{"10.0.248.0", "10.0.248.1", "10.0.255.254", "10.0.255.255"}}

func main() {
	mode := flag.String("mode", "normal", "'cleanup' or 'normal' (default)")

	flag.Parse()

	err, AWS_ACCESS_KEY_ID, stderr := aws.Execute("aws configure get aws_access_key_id", true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	err, AWS_ACCESS_SECRET, stderr := aws.Execute("aws configure get aws_secret_access_key", true, false)

	// err, stdout, stderr = aws.Execute("aws rds help", true, false)
	// log.Printf("%v %s %s", err, out, _err)

	fmt.Println("AWS_ACCESS_KEY_ID:", AWS_ACCESS_KEY_ID)
	fmt.Println("AWS_ACCESS_SECRET:", AWS_ACCESS_SECRET)

	awsm := &aws.AWS{REGION}

	if *mode == "cleanup" {
		doCleanup(awsm)
		os.Exit(0)
	}

	if *mode != "normal" {
		fmt.Println("Unknown mode.", *mode)
		os.Exit(1)
	}

	err, stack := awsm.GetStackWithTag("alpha.eksctl.io/cluster-name", CLUSTER_KEY)

	if stack != nil && stack.StackStatus == "CREATE_COMPLETE" {
		fmt.Println("Stack already exists. Exiting.")
	}

	// wait while creating

	// check db
	if stack != nil && stack.StackStatus == "CREATE_COMPLETE" {
		err, rds := GetRds(awsm)

		if err != nil {
			fmt.Println("err:", err)
		}

		fmt.Println("rds:", rds)
	}

	err, vpcs := awsm.GetVPCs()

	if err != nil {
		fmt.Println("err:", err)
	}

	vpc := getVpc(vpcs)

	err, availabilityZones := awsm.GetAvailabilityZones()

	var zones = []string{}

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("availabilityZones:", availabilityZones)

	for _, availabilityZone := range availabilityZones {
		zones = append(zones, availabilityZone.ZoneName)
	}

	if vpc != nil {
		fmt.Println("Found vpc:", vpc)

	} else {
		fmt.Printf("Unable to find vpc with name %s-vpc. Creating ...", CLUSTER_KEY)

		err, vpc = createVPC(awsm)

		if err != nil {
			fmt.Println("err:", err)
		}
	}

	VPC_ID = vpc.VpcId

	err, subnets := awsm.GetSubnetsByVPCId(vpc.VpcId)

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("subnets:", subnets)

	if len(subnets) == 0 {
		err, subnets = createSubnet(awsm, *vpc, zones)

		if err != nil {
			fmt.Println("err:", err)
		}

		fmt.Println("subnets:", subnets)
	}

	err, internetGateway := getInternetGateway(awsm)

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("internetGateway:", internetGateway)

	if internetGateway == nil {
		err, igw := createInternetGateway(awsm, *vpc)

		if err != nil {
			fmt.Println("err:", err)
		}

		fmt.Println("igw:", igw)
	}

	err, addresses := getAddresses(awsm)

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("addresses:", addresses)

	if len(addresses) < len(availabilityZones) {
		for _, _ = range make([]int, len(availabilityZones)-len(addresses)) {
			err, address := awsm.AllocateAddress()

			if err != nil {
				fmt.Println("err:", err)
			}

			fmt.Println("address:", address)
		}
	}

	err, natGateways := getNatGateways(awsm)

	if err != nil {
		fmt.Println("err:", err)
	}

	fmt.Println("natGateways:", natGateways)

	if len(natGateways) < len(availabilityZones) {
		err, ngw := createNatGateway(awsm, subnets, addresses)

		if err != nil {
			fmt.Println("err:", err)
		}

		fmt.Println("ngw:", ngw)
	}

	analyseRoutes(awsm, *vpc, subnets)

	// deployDatabase(awsm, subnets)

	// createStack(awsm, subnets)
}
