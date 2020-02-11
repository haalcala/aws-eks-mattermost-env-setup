package aws

// IAM

type IAMListRolesResponse struct {
	Roles []AWSIAMRoleType `json:"Roles"`
}

// EKS

type EKSListFargateProfilesResponse struct {
	FargateProfileNames []string `json:"fargateProfileNames"`
}

type EKSDescribeFargateProfileResponse struct {
	FargateProfile AWSEKSFargateProfileType `json:"fargateProfile"`
}

type EKSCreateFargateProfileResponse struct {
	FargateProfile AWSEKSFargateProfileType `json:"fargateProfile"`
}

type EKSListClustersResponse struct {
	Clusters []string `json:"clusters"`
}

type EKSDeleteClusterResponse struct {
	Cluster AWSEKSClusterType `json:"cluster"`
}

type EKSDescribeClusterResponse struct {
	Cluster AWSEKSClusterType `json:"cluster"`
}

type DescribeVPCResponse struct {
	Vpcs []AWSVPCType `json:"Vpcs"`
}

type RDSCreateDBSubnetGroupResponse struct {
	DBSubnetGroup AWSRDSSubnetGroupType `json:"DBSubnetGroup"`
}

// Subnets

type EC2DescribeSubnetsResponse struct {
	Subnets []AWSSubnetType `json:"Subnets"`
}

type EC2CreateSubnetResponse struct {
	Subnet AWSSubnetType `json:"Subnet"`
}

// RouteTables

type EC2DescribeRouteTablesResponse struct {
	RouteTables []AWSRouteTableType `json:"RouteTables"`
}

type EC2CreateRouteTableResponse struct {
	RouteTable AWSRouteTableType `json:"RouteTable"`
}

type EC2AssociateRouteTableResponse struct {
	AssociationId string `json:"AssociationId"`
}

type EC2CreateRouteResponse struct {
	Return bool `json:"Return"`
}

// RDS

type RDSCreateDBInstanceResponse struct {
	DBInstance AWSDBInstanceType `json:"DBInstance"`
}

type RDSDescribeDBInstancesResponse struct {
	DBInstances []AWSDBInstanceType `json:"DBInstances"`
}

type RDSDescribeSubnetGroupsResponse struct {
	DBSubnetGroups []AWSRDSSubnetGroupType `json:"DBSubnetGroups"`
}

// Internet Gateways

type AWSEC2DescribeInternetGatewaysResponse struct {
	InternetGateways []AWSInternetGatewayType `json:"InternetGateways"`
}

type AWSEC2CreateInternetGatewayResponse struct {
	InternetGateway *AWSInternetGatewayType `json:"InternetGateway"`
}

// NAT Gateways

type AWSEC2CreateNatGatewayResponse struct {
	NatGateway *AWSNatGatewayType `json:"NatGateway"`
}

type AWSEC2DescribeNatGatewaysResponse struct {
	NatGateways []AWSNatGatewayType `json:"NatGateways"`
}

// Availability Zones

type EC2DescribeAvailabilityZonesReponse struct {
	AvailabilityZones []AWSAvailabilityZoneType `json:"AvailabilityZones"`
}

// VPC

type EC2CreateVPCResponse struct {
	Vpc *AWSVPCType `json:"Vpc"`
}

// IAM Types

type AWSIAMRoleType struct {
	Description              string                                 `json:"Description"`
	RoleId                   string                                 `json:"RoleId"`
	CreateDate               string                                 `json:"CreateDate"`
	RoleName                 string                                 `json:"RoleName"`
	Path                     string                                 `json:"Path"`
	Arn                      string                                 `json:"Arn"`
	AssumeRolePolicyDocument AWSIAMRoleAssumeRolePolicyDocumentType `json:"AssumeRolePolicyDocument"`
	MaxSessionDuration       int                                    `json:"MaxSessionDuration"`
}

type AWSIAMRoleAssumeRolePolicyDocumentType struct {
	Version   string                                            `json:"Version"`
	Statement []AWSIAMRoleAssumeRolePolicyDocumentStatementType `json:"Statement"`
}

type AWSIAMRoleAssumeRolePolicyDocumentStatementType struct {
	Action string `json:"Action"`
	Effect string `json:"Effect"`
	Sid    string `json:"Sid"`
	// "Principal": {
	// 	"Service": [
	// 		"lambda.amazonaws.com",
	// 		"apigateway.amazonaws.com"
	// 	]
	// },
}

// Addresses

type AWSEC2DescribeAddressesResponse struct {
	Addresses []AWSAddressType `json:"Addresses"`
}

type AWSEC2AllocateAddressResponse struct {
	Domain         string `json:"Domain"`
	PublicIpv4Pool string `json:"PublicIpv4Pool"`
	PublicIp       string `json:"PublicIp"`
	AllocationId   string `json:"AllocationId"`
}

type AWSAddressType struct {
	Domain                  string  `json:"Domain"`
	PublicIpv4Pool          string  `json:"PublicIpv4Pool"`
	InstanceId              string  `json:"InstanceId"`
	NetworkInterfaceId      string  `json:"NetworkInterfaceId"`
	AssociationId           *string `json:"AssociationId"`
	NetworkInterfaceOwnerId string  `json:"NetworkInterfaceOwnerId"`
	PublicIp                string  `json:"PublicIp"`
	AllocationId            *string `json:"AllocationId"`
	PrivateIpAddress        string  `json:"PrivateIpAddress"`
}

type AWSEKSClusterType struct {
	Name               string                              `json:"name"`
	Arn                string                              `json:"arn"`
	Version            string                              `json:"version"`
	Endpoint           string                              `json:"endpoint"`
	RoleArn            string                              `json:"roleArn"`
	Status             string                              `json:"status"`
	ResourcesVpcConfig AWSEKSClusterResourcesVPCConfigType `json:"resourcesVpcConfig"`
	CreatedAt          int                                 `json:"createdAt"`
	// "logging": {
	// 	"clusterLogging": [
	// 		{
	// 			"types": [
	// 				"api",
	// 				"audit",
	// 				"authenticator",
	// 				"controllerManager",
	// 				"scheduler"
	// 			],
	// 			"enabled": false
	// 		}
	// 	]
	// },
	// "identity": {
	// 	"oidc": {
	// 		"issuer": "https://oidc.eks.us-east-2.amazonaws.com/id/59E19E3C209A20C055EB04F9E732C135"
	// 	}
	// },
	// "certificateAuthority": {
	// 	"data": ""
	// },
	// "platformVersion": "eks.7",
	// "tags": {}
}

type AWSEKSClusterResourcesVPCConfigType struct {
	subnetIds              []string `json:"subnetIds"`
	securityGroupIds       []string `json:"securityGroupIds"`
	publicAccessCidrs      []string `json:"publicAccessCidrs"`
	clusterSecurityGroupId string   `json:"clusterSecurityGroupId"`
	vpcId                  string   `json:"vpcId"`
	endpointPublicAccess   bool     `json:"endpointPublicAccess"`
	endpointPrivateAccess  bool     `json:"endpointPrivateAccess"`
}

type AWSEKSFargateProfileType struct {
	FargateProfileName  string                             `json:"fargateProfileName"`
	FargateProfileArn   string                             `json:"fargateProfileArn"`
	ClusterName         string                             `json:"clusterName"`
	PodExecutionRoleArn string                             `json:"podExecutionRoleArn"`
	Status              string                             `json:"status"`
	CreatedAt           int                                `json:"createdAt"`
	Subnets             []AWSEKSFargateProfileSelectorType `json:"subnets"`
	Selectors           []string                           `json:"selectors"`
	// "tags": {}
}

type AWSEKSFargateProfileSelectorType struct {
	Namespace string `json:"namespace"`
}

type AWSNatGatewayType struct {
	NatGatewayAddresses []AWSNatGatewayAddressType `json:"NatGatewayAddresses"`
	VpcId               string                     `json:"VpcId"`
	Tags                []TagType                  `json:"Tags"`
	State               string                     `json:"State"`
	NatGatewayId        string                     `json:"NatGatewayId"`
	SubnetId            string                     `json:"SubnetId"`
	CreateTime          string                     `json:"CreateTime"`
}

type AWSNatGatewayAddressType struct {
	PublicIp           string `json:"PublicIp"`
	NetworkInterfaceId string `json:"NetworkInterfaceId"`
	AllocationId       string `json:"AllocationId"`
	PrivateIp          string `json:"PrivateIp"`
}

type AWSInternetGatewayType struct {
	Attachments       []AWSInternetGatewayAttachmentType `json:"Attachments"`
	InternetGatewayId string                             `json:"InternetGatewayId"`
	OwnerId           string                             `json:"OwnerId"`
	Tags              []TagType                          `json:"Tags"`
}

type AWSInternetGatewayAttachmentType struct {
	State string `json:"State"`
	VpcId string `json:"VpcId"`
}

type AWSVPCType struct {
	CidrBlock                   string    `json:"CidrBlock"`
	DhcpOptionsId               string    `json:"DhcpOptionsId"`
	State                       string    `json:"State"`
	VpcId                       string    `json:"VpcId"`
	OwnerId                     string    `json:"OwnerId"`
	InstanceTenancy             string    `json:"InstanceTenancy"`
	Ipv6CidrBlockAssociationSet string    `json:"Ipv6CidrBlockAssociationSet"`
	CidrBlockAssociationSet     string    `json:"CidrBlockAssociationSet"`
	IsDefault                   string    `json:"IsDefault"`
	Tags                        []TagType `json:"Tags"`
}

type AWSSubnetType struct {
	AvailabilityZone            string    `json:"AvailabilityZone"`
	AvailabilityZoneId          string    `json:"AvailabilityZoneId"`
	AvailableIpAddressCount     string    `json:"AvailableIpAddressCount"`
	CidrBlock                   string    `json:"CidrBlock"`
	DefaultForAz                string    `json:"DefaultForAz"`
	MapPublicIpOnLaunch         string    `json:"MapPublicIpOnLaunch"`
	State                       string    `json:"State"`
	SubnetId                    string    `json:"SubnetId"`
	VpcId                       string    `json:"VpcId"`
	OwnerId                     string    `json:"OwnerId"`
	AssignIpv6AddressOnCreation string    `json:"AssignIpv6AddressOnCreation"`
	Ipv6CidrBlockAssociationSet string    `json:"Ipv6CidrBlockAssociationSet"`
	SubnetArn                   string    `json:"SubnetArn"`
	Tags                        []TagType `json:"Tags"`
}

// Routes

type AWSRouteTableAssociationType struct {
	Main                    bool   `json:"Main"`
	RouteTableAssociationId string `json:"RouteTableAssociationId"`
	RouteTableId            string `json:"RouteTableId"`
	SubnetId                string `json:"SubnetId"`
}

type AWSRouteTableType struct {
	RouteTableId string                         `json:"RouteTableId"`
	Routes       []AWSRouteType                 `json:"Routes"`
	VpcId        string                         `json:"VpcId"`
	Associations []AWSRouteTableAssociationType `json:"Associations"`
	Tags         []TagType                      `json:"Tags"`
}

type AWSRouteType struct {
	DestinationCidrBlock string `json:"DestinationCidrBlock"`
	GatewayId            string `json:"GatewayId"`
	Origin               string `json:"Origin"`
	State                string `json:"State"`
}

type AWSAvailabilityZoneType struct {
	State              string `json:"State"`
	OptInStatus        string `json:"OptInStatus"`
	Messages           string `json:"Messages"`
	RegionName         string `json:"RegionName"`
	ZoneName           string `json:"ZoneName"`
	ZoneId             string `json:"ZoneId"`
	GroupName          string `json:"GroupName"`
	NetworkBorderGroup string `json:"NetworkBorderGroup"`
}

type RDSDescribeResponse struct {
	DBInstances []RDSPayload `json:"DBInstances"`
}

type RDSPayload struct {
	DBInstanceIdentifier string      `json:"DBInstanceIdentifier"`
	DBInstanceStatus     string      `json:"DBInstanceStatus"`
	NotExisting          interface{} `json:"NotExisting"`
}

type CFDescribeStacksResponse struct {
	Stacks []AWSCFStackType `json:"Stacks"`
}

type AWSCFStackType struct {
	StackId     string    `json:"StackId"`
	StackName   string    `json:"StackName"`
	Tags        []TagType `json:"Tags"`
	StackStatus string    `json:"StackStatus"`
}

type TagType struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type Subnet struct {
	CIDR, Start, End, Netmask string
}

type AWSDBInstanceType struct {
	AllocatedStorage      int `json:"AllocatedStorage"`
	BackupRetentionPeriod int `json:"BackupRetentionPeriod"`
	DbInstancePort        int `json:"DbInstancePort"`
	MonitoringInterval    int `json:"MonitoringInterval"`

	DBSecurityGroups                 string `json:"DBSecurityGroups"`
	ReadReplicaDBInstanceIdentifiers string `json:"ReadReplicaDBInstanceIdentifiers"`
	DBInstanceIdentifier             string `json:"DBInstanceIdentifier"`
	DBInstanceClass                  string `json:"DBInstanceClass"`
	Engine                           string `json:"Engine"`
	DBInstanceStatus                 string `json:"DBInstanceStatus"`
	MasterUsername                   string `json:"MasterUsername"`
	PreferredBackupWindow            string `json:"PreferredBackupWindow"`
	PreferredMaintenanceWindow       string `json:"PreferredMaintenanceWindow"`
	EngineVersion                    string `json:"EngineVersion"`
	LicenseModel                     string `json:"LicenseModel"`
	StorageType                      string `json:"StorageType"`
	DbiResourceId                    string `json:"DbiResourceId"`
	CACertificateIdentifier          string `json:"CACertificateIdentifier"`
	DBInstanceArn                    string `json:"DBInstanceArn"`

	MultiAZ                          bool `json:"MultiAZ"`
	AutoMinorVersionUpgrade          bool `json:"AutoMinorVersionUpgrade"`
	PubliclyAccessible               bool `json:"PubliclyAccessible"`
	StorageEncrypted                 bool `json:"StorageEncrypted"`
	DomainMemberships                bool `json:"DomainMemberships"`
	CopyTagsToSnapshot               bool `json:"CopyTagsToSnapshot"`
	IAMDatabaseAuthenticationEnabled bool `json:"IAMDatabaseAuthenticationEnabled"`
	PerformanceInsightsEnabled       bool `json:"PerformanceInsightsEnabled"`

	// "VpcSecurityGroups": [
	// 	{
	// 		"VpcSecurityGroupId": "sg-f839b688",
	// 		"Status": "active"
	// 	}
	// ],
	// "DBParameterGroups": [
	// 	{
	// 		"DBParameterGroupName": "default.mysql5.6",
	// 		"ParameterApplyStatus": "in-sync"
	// 	}
	// ],
	DBSubnetGroup          AWSRDSSubnetGroupType              `json:"DBSubnetGroup"`
	PendingModifiedValues  AWSRDSPendingModifiedValuesType    `json:"PendingModifiedValues"`
	OptionGroupMemberships []AWSRDSOptionGroupMembershipsType `json:"OptionGroupMemberships"`
}

type AWSRDSOptionGroupMembershipsType struct {
	OptionGroupName string `json:"OptionGroupName"`
	Status          string `json:"Status"`
}

type AWSRDSDBSubnetGroupSubnetType struct {
	SubnetIdentifier       string                                              `json:"SubnetIdentifier"`
	SubnetStatus           string                                              `json:"SubnetStatus"`
	SubnetAvailabilityZone AWSRDSDBSubnetGroupSubnetSubnetAvailabilityZoneType `json:"SubnetAvailabilityZone"`
}

type AWSRDSDBSubnetGroupSubnetSubnetAvailabilityZoneType struct {
	Name string `json:"Name"`
}

type AWSRDSPendingModifiedValuesType struct {
	MasterUserPassword           string                                                      `json:"MasterUserPassword"`
	PendingCloudwatchLogsExports AWSRDSPendingModifiedValuesPendingCloudwatchLogsExportsType `json:"PendingCloudwatchLogsExports"`
}

type AWSRDSPendingModifiedValuesPendingCloudwatchLogsExportsType struct {
	LogTypesToEnable []string `json:"LogTypesToEnable"`
}

type AWSRDSSubnetGroupType struct {
	DBSubnetGroupName        string                        `json:"DBSubnetGroupName"`
	DBSubnetGroupDescription string                        `json:"DBSubnetGroupDescription"`
	VpcId                    string                        `json:"VpcId"`
	SubnetGroupStatus        string                        `json:"SubnetGroupStatus"`
	DBSubnetGroupArn         string                        `json:"DBSubnetGroupArn"`
	Subnets                  []AWSRDSSubnetGroupSubnetType `json:"Subnets"`
	Tags                     []TagType                     `json:"Tags"`
}

type AWSRDSSubnetGroupSubnetType struct {
	SubnetIdentifier       string                                      `json:"SubnetIdentifier"`
	SubnetStatus           string                                      `json:"SubnetStatus"`
	SubnetAvailabilityZone AWSRDSSubnetGroupSubnetAvailabilityZoneType `json:"SubnetAvailabilityZone"`
}

type AWSRDSSubnetGroupSubnetAvailabilityZoneType struct {
	Name string `json:"Name"`
}
