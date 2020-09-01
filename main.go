package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/bobziuchkovski/digest"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ProvisionJson struct {
	Certificate    Certificate    `json:"certificate"`
	ConnectionInfo ConnectionInfo `json:"connection_info"`
	Topics         Topics         `json:"topics"`
	Services       Services       `json:"services"`
}

type Certificate struct {
	CaCert     string `json:"ca_cert"`
	SignedCert string `json:"signed_cert"`
}

type ConnectionInfo struct {
	ClientId           string             `json:"client_id"`
	Mqtt               Mqtt               `json:"mqtt"`
	CredentialProvider CredentialProvider `json:"CredentialProvider"`
}

type Mqtt struct {
	AlpnPort int    `json:"alpn_port"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
}

type CredentialProvider struct {
	EndpointAddress string `json:"endpoint_address"`
	RoleAlias       string `json:"role_alias"`
}

type Topics struct {
	Publish   Publish   `json:"publish"`
	Subscribe Subscribe `json:"subscribe"`
}

type Publish struct {
	Data string `json:"data"`
	Logs string `json:"logs"`
	Ping string `json:"ping"`
}

type Subscribe struct {
	Pong string `json:"pong"`
}

type Services struct {
	S3 S3 `json:"s3"`
}

type S3 struct {
	BucketArn string `json:"bucket_arn"`
	BucketId  string `json:"bucket_id"`
}

type VeraDeviceCommand struct {
	Device int `json:"Device"`
	State int `json:"State"`
}

type AssumeRoleWithCertificate struct {
	Credentials TemporaryCredentials `json:"credentials"`
}

type TemporaryCredentials struct {
	AccessKeyID string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	SessionToken string `json:"sessionToken"`
	Expiration string `json:"expiration"`
}

var HttpClient *http.Client
var HttpsClient *http.Client
var ThingConfig ProvisionJson
var TLSConfig *tls.Config

const (
	MaxIdleConnections int = 20
	RequestTimeout     int = 5
)

func init() {
	HttpClient = createHTTPClient()
}

// createHTTPClient for connection re-use
func createHTTPClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}

	return client
}

func createHTTPSClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: MaxIdleConnections,
			TLSClientConfig: TLSConfig,
		},
		Timeout: time.Duration(RequestTimeout) * time.Second,
	}

	return client
}

func veraControllerCommand (deviceNum,newTargetValue int ) error {

	verabody := fmt.Sprintf("id=action&output_format=json&serviceId=urn:upnp-org:serviceId:SwitchPower1&action=SetTarget&newTargetValue=%d&DeviceNum=%d", newTargetValue,deviceNum)

	fmt.Println(verabody)

	endPoint := "http://vera.bovee.io:3480/data_request"

	req, err := http.NewRequest("POST", endPoint, bytes.NewBuffer([]byte(verabody)))
	if err != nil {
		log.Fatalf("Error Occured. %+v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := HttpClient.Do(req)
	if err != nil && response == nil {
		log.Fatalf("Error sending request to API endpoint. %+v", err)
		return err
	}

	// Close the connection to reuse it
	defer response.Body.Close()

	// Let's check if the work actually is done
	// We have seen inconsistencies even when we get 200 OK response
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Couldn't parse response body. %+v", err)
		return err
	}

	log.Println("Response Body:", string(body))
	return nil
}

func getPrivateKey(env, path string) (string, error) {
	var err error

	// prefer environment variable if it exists
	envVar := fmt.Sprintf("%s_IOT_PRIVATE_KEY", strings.ToUpper(env))
	if os.Getenv(envVar) != "" {
		return os.Getenv(envVar), err
	}

	// try to load from file
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return string(file), nil
}

func NewTLSConfig(caPem string, certificatePem string, privateKeyPem string) *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()

	// load AWS IoT Core CA cert from provision.json
	certpool.AppendCertsFromPEM([]byte(caPem))

	// Import client certificate from provision.json and key from environment
	cert, err := tls.X509KeyPair([]byte(certificatePem), []byte(privateKeyPem))
	if err != nil {
		panic(err)
	}

	// Just to print out the client certificate..
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}
	//fmt.Println(cert.Leaf)

	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		Certificates: []tls.Certificate{cert},
	}
}

func captureImage() (string, error) {

	fmt.Println("Waiting...")
	time.Sleep(5 * time.Second)
	fmt.Println("Done")

	t := digest.NewTransport(os.Getenv("AXIS"), os.Getenv("AXISPWD"))
	c, err := t.Client()
	if err != nil {
		return "", err
	}
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go
	resp, err := c.Get("http://192.168.1.41/axis-cgi/jpg/image.cgi")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fmt.Println(resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	currentTime := time.Now()
	filename :=  fmt.Sprintf("%s.jpg", currentTime.Format("2006-01-02-15:04:05.000000000"))
	fmt.Printf("Writing file: %s\n", filename)
	err = ioutil.WriteFile(fmt.Sprintf("/config/%s",filename), body, 0666)
	if err != nil {
		return "", err
	}

	fmt.Println("Done")

	return filename, nil
}

type LogMessage struct {
	LogLine string `json:"LogLine"`
}

func sendLogMessage(client MQTT.Client, line string) {
	logline := &LogMessage{
		LogLine: line,
	}
	mqttJson, _ := json.Marshal(logline)
	client.Publish(ThingConfig.Topics.Publish.Logs, 0, false, mqttJson)
}

func onMessageReceived(client MQTT.Client, message MQTT.Message) {
	fmt.Printf("Received \"%s\" on topic: %s\n", message.Payload(), message.Topic())

	var veraDeviceCommand VeraDeviceCommand
	err := json.Unmarshal(message.Payload(), &veraDeviceCommand)
	if err != nil {
		panic(err)
	}

	fmt.Println(veraDeviceCommand.Device)
	fmt.Println(veraDeviceCommand.State)

	err = veraControllerCommand(veraDeviceCommand.Device,veraDeviceCommand.State)
	if err != nil {
		panic(err)
	} else {
		fmt.Println("Sent command to vera")
	}

	filename, err := captureImage()
	if err != nil {
		panic(err)
	}

	fmt.Println("Captured Image")

	err = uploadToS3(filename)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Processed %s", filename)
	sendLogMessage(client,fmt.Sprintf("Processed %s from %s\n", filename, ThingConfig.ConnectionInfo.ClientId))

}

func processProvisionJson(jsonFilePtr string) (ProvisionJson, error) {
	var provisionJson ProvisionJson
	jsonFile, err := os.Open(jsonFilePtr)
	// if we os.Open returns an error then handle it
	if err != nil {
		return provisionJson, err
	}

	// defer the closing of our jsonFile so that we can parse it later on
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &provisionJson)
	if err != nil {
		return provisionJson, err
	}

	return provisionJson, err
}

func getTemporaryCredentials () (AssumeRoleWithCertificate, error) {

	var temporaryCredentials AssumeRoleWithCertificate


	credentialUrl := fmt.Sprintf("https://%s/role-aliases/%s/credentials",ThingConfig.ConnectionInfo.CredentialProvider.EndpointAddress,ThingConfig.ConnectionInfo.CredentialProvider.RoleAlias)
	req, err := http.NewRequest("GET", credentialUrl, nil)
	if err != nil {
		return temporaryCredentials, err
	}
	req.Header.Set("X-Amzn-Iot-Thingname", ThingConfig.ConnectionInfo.ClientId)

	resp, err := HttpsClient.Do(req)
	if err != nil {
		return temporaryCredentials, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return temporaryCredentials, err
		}

		// decode the response
		err = json.Unmarshal(bodyBytes, &temporaryCredentials)
		if err != nil {
			return temporaryCredentials, err
		}
	}

	return temporaryCredentials, nil
}

func main() {
	jsonFilePtr := flag.String("conf", "", "path to provision.json file")
	env := flag.String("env", "", "the environment for the devices")
	pathPtr := flag.String("key", "", "the path to the private key.  prefers `IOT_PRIVATE_KEY` env var")

	flag.Parse()

	// check if -conf was passed
	if *jsonFilePtr == "" {
		fmt.Println("Flags not passed correctly.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *env == "" {
		*env = "dev"
	}

	if *env != "dev" && *env != "stage" && *env != "prod" {
		//fmt.Println(fmt.Sprintf("env not properly passed, must be one of : dev, stage, prod. got `%s`", *env))
		fmt.Printf("env not properly passed, must be one of : dev, stage, prod. got `%s`\n\n\n", *env)
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *pathPtr == "" && os.Getenv(fmt.Sprintf("%s_IOT_PRIVATE_KEY", strings.ToUpper(*env))) == "" {
		fmt.Println("Flags not passed correctly.")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	ThingConfig, err = processProvisionJson(*jsonFilePtr)
	if err != nil {
		panic(err)
	}

	privateKeyPem, err := getPrivateKey(*env, *pathPtr)
	if err != nil {
		panic(err)
	}

	// set up TLS
	TLSConfig = NewTLSConfig(ThingConfig.Certificate.CaCert, ThingConfig.Certificate.SignedCert, privateKeyPem)

	// set up credential provider http client
	HttpsClient = createHTTPSClient()

	// set up broker stuff
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// configure broker
	opts := MQTT.NewClientOptions()
	opts.AddBroker("ssl://" + ThingConfig.ConnectionInfo.Mqtt.Host + ":" + strconv.Itoa(ThingConfig.ConnectionInfo.Mqtt.Port))
	opts.SetClientID(ThingConfig.ConnectionInfo.ClientId).SetTLSConfig(TLSConfig)
	opts.OnConnect = func(c MQTT.Client) {
		if token := c.Subscribe(ThingConfig.Topics.Subscribe.Pong, 0, onMessageReceived); token.Wait() && token.Error() != nil {
			panic(token.Error())
		}
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}else {
		fmt.Printf("Connected to %s\n", *env)
	}

	<-c

}

func uploadToS3 (filename string) error {
	awscreds, err := getTemporaryCredentials()
	if err != nil {
		return err
	}

	filepath := fmt.Sprintf("/config/%s", filename)
	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	//fmt.Println(awscreds.Credentials.AccessKeyID)
	//fmt.Println(awscreds.Credentials.SecretAccessKey)
	//fmt.Println(awscreds.Credentials.SessionToken)
	//fmt.Println(awscreds.Credentials.Expiration)

	fmt.Println("------------")
	fmt.Printf("Upload %s to S3\n", filename)
	fmt.Println("------------")
	fmt.Println("------------")

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(awscreds.Credentials.AccessKeyID, awscreds.Credentials.SecretAccessKey, awscreds.Credentials.SessionToken),
	})
	if err != nil {
		return err
	}

	svc := s3manager.NewUploader(sess)
	fmt.Println("Uploading file to S3...")
	result, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(ThingConfig.Services.S3.BucketId),
		Key:    aws.String(fmt.Sprintf("%s/%s", ThingConfig.ConnectionInfo.ClientId,filename)),
		Body:   file,
	})
	if err != nil {
		return err
	}

	fmt.Println(result)

	err = os.Remove(filepath)
	if err != nil {
		return err
	}

	fmt.Println("File Deleted")

	return nil
}

