package main

import (
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"
)

func TestNewTLSConfig(t *testing.T) {
	type args struct {
		caPem          string
		certificatePem string
		privateKeyPem  string
	}
	tests := []struct {
		name string
		args args
		want *tls.Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTLSConfig(tt.args.caPem, tt.args.certificatePem, tt.args.privateKeyPem); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTLSConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_captureImage(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := captureImage()
			if (err != nil) != tt.wantErr {
				t.Errorf("captureImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("captureImage() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createHTTPClient(t *testing.T) {
	tests := []struct {
		name string
		want *http.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createHTTPClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createHTTPClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createHTTPSClient(t *testing.T) {
	tests := []struct {
		name string
		want *http.Client
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createHTTPSClient(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createHTTPSClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getPrivateKey(t *testing.T) {
	type args struct {
		env  string
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getPrivateKey(tt.args.env, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("getPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getPrivateKey() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTemporaryCredentials(t *testing.T) {
	tests := []struct {
		name    string
		want    AssumeRoleWithCertificate
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTemporaryCredentials()
			if (err != nil) != tt.wantErr {
				t.Errorf("getTemporaryCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getTemporaryCredentials() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//func Test_onMessageReceived(t *testing.T) {
//	type args struct {
//		client  mqtt.Client
//		message mqtt.Message
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//		})
//	}
//}

func Test_processProvisionJson(t *testing.T) {
	type args struct {
		jsonFilePtr string
	}
	tests := []struct {
		name    string
		args    args
		want    ProvisionJson
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := processProvisionJson(tt.args.jsonFilePtr)
			if (err != nil) != tt.wantErr {
				t.Errorf("processProvisionJson() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("processProvisionJson() got = %v, want %v", got, tt.want)
			}
		})
	}
}

//func Test_sendLogMessage(t *testing.T) {
//	type args struct {
//		client mqtt.Client
//		line   string
//	}
//	tests := []struct {
//		name string
//		args args
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//		})
//	}
//}

func Test_uploadToS3(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := uploadToS3(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("uploadToS3() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_veraControllerCommand(t *testing.T) {
	type args struct {
		deviceNum      int
		newTargetValue int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := veraControllerCommand(tt.args.deviceNum, tt.args.newTargetValue); (err != nil) != tt.wantErr {
				t.Errorf("veraControllerCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}