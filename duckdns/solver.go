package duckdns

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/cert-manager/cert-manager/pkg/acme/webhook"
	"github.com/cert-manager/cert-manager/pkg/acme/webhook/apis/acme/v1alpha1"
	"github.com/cert-manager/cert-manager/pkg/issuer/acme/dns/util"

	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"

	duckdnsgo "github.com/ebrianne/duckdns-go/duckdns"
)

func NewSolver() webhook.Solver {
	return &duckDNSProviderSolver{}
}

type duckDNSProviderSolver struct {
	client *kubernetes.Clientset
}

func (s *duckDNSProviderSolver) Name() string {
	return "duckdns"
}

func (s *duckDNSProviderSolver) validateConfig(cfg *Config) error {
	if cfg.APITokenSecretRef.LocalObjectReference.Name == "" {
		return errors.New("no api token secret provided in DuckDNS config")
	}
	return nil
}

func (s *duckDNSProviderSolver) newClientFromChallenge(ch *v1alpha1.ChallengeRequest) (*duckdnsgo.Client, error) {
	cfg, err := loadConfig(ch.Config)
	if err != nil {
		return nil, err
	}

	if err := s.validateConfig(&cfg); err != nil {
		return nil, err
	}

	klog.Infof("Decoded config: %v", cfg)

	apiToken, err := s.getApiToken(&cfg, ch.ResourceNamespace)
	if err != nil {
		return nil, fmt.Errorf("get credential error: %v", err)
	}

	config := &duckdnsgo.Config{
		Token:       *apiToken,
		DomainNames: s.getDNSName(ch),
	}
	client := duckdnsgo.NewClient(http.DefaultClient, config)

	return client, nil
}

func (s *duckDNSProviderSolver) getDNSName(ch *v1alpha1.ChallengeRequest) []string {
	dnsFromChallenge := ch.DNSName
	klog.Infof("DNSName from ChallengeRequest is %v", dnsFromChallenge)

	duckdnszone := strings.TrimSuffix(dnsFromChallenge, "duckdns.org")
	duckdnszone = util.UnFqdn(duckdnszone)
	split := strings.Split(duckdnszone, ".")

	if len(split) == 2 {
		klog.Infof("Got dns domain from challenge %v", split[1])
		return []string{split[1]}
	}

	klog.Infof("Got dns domain from challenge %v", duckdnszone)
	return []string{duckdnszone}
}

func (s *duckDNSProviderSolver) getApiToken(cfg *Config, namespace string) (*string, error) {
	secretName := cfg.APITokenSecretRef.LocalObjectReference.Name

	secret, err := s.client.CoreV1().Secrets(namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load secret %q", namespace+"/"+secretName)
	}

	data, ok := secret.Data[cfg.APITokenSecretRef.Key]
	if !ok {
		return nil, fmt.Errorf("key %q not found in secret \"%s/%s\"", cfg.APITokenSecretRef.Key, secretName, namespace)
	}

	apiKey := string(data)
	return &apiKey, nil
}

func (s *duckDNSProviderSolver) Present(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Presenting txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	client, err := s.newClientFromChallenge(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}

	domain := client.Config.DomainNames[0]
	klog.Infof("Present txt record for domain %v", domain)

	if _, err := client.UpdateRecord(context.Background(), ch.Key); err != nil {
		klog.Errorf("Add txt record %q error: %v", ch.ResolvedFQDN, err)
		return err
	}

	klog.Infof("Presented txt record %v", ch.ResolvedFQDN)
	return nil
}

func (s *duckDNSProviderSolver) CleanUp(ch *v1alpha1.ChallengeRequest) error {
	klog.Infof("Cleaning up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	client, err := s.newClientFromChallenge(ch)
	if err != nil {
		klog.Errorf("New client from challenge error: %v", err)
		return err
	}

	domain := client.Config.DomainNames[0]
	klog.Infof("Cleaning up txt record for domain %v", domain)

	record, err := client.GetRecord()
	if err != nil {
		klog.Errorf("Get txt record %v error: %v", ch.ResolvedFQDN, err)
		return err
	}
	klog.Infof("Got txt record: %v", record)

	if record != ch.Key {
		klog.Errorf("Record value %v does not match key %v for %v", record, ch.Key, ch.ResolvedFQDN)
		return errors.New("record value does not match")
	}

	if _, err := client.ClearRecord(context.Background(), ch.Key); err != nil {
		klog.Errorf("Delete domain record %v error: %v", ch.ResolvedFQDN, err)
		return err
	}

	klog.Infof("Cleaned up txt record: %v %v", ch.ResolvedFQDN, ch.ResolvedZone)
	return nil
}

func (s *duckDNSProviderSolver) Initialize(kubeClientConfig *rest.Config, stopCh <-chan struct{}) error {
	cl, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return err
	}

	s.client = cl
	return nil
}
