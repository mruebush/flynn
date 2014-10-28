package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	cc "github.com/flynn/flynn/controller/client"
	ct "github.com/flynn/flynn/controller/types"
	"github.com/flynn/flynn/discoverd/client"
	"github.com/flynn/flynn/pkg/resource"
	"github.com/flynn/flynn/router/types"
)

type generator struct {
	conf        *config
	client      *cc.Client
	resourceIds map[string]string
}

type example struct {
	name string
	f    func()
}

func main() {
	conf, err := loadConfigFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	err = discoverd.Connect(conf.controllerDomain + ":1111")
	if err != nil {
		log.Fatal(err)
	}

	client, err = cc.NewClient("http://"+conf.controllerDomain, conf.controllerKey)
	if err != nil {
		log.Fatal(err)
	}
	client.HTTP.Transport = &roundTripRecorder{roundTripper: &http.Transport{}}

	e := &generator{
		conf:        conf,
		client:      client,
		resourceIds: make(map[string]string),
	}

	providerLog := log.New(os.Stdout, "provider: ", 1)
	go e.listenAndServe(providerLog)

	examples := []example{
		{"key_create", e.createKey},
		{"key_get", e.getKey},
		{"key_list", e.listKeys},
		{"key_delete", e.deleteKey},
		{"app_create", e.createApp},
		{"app_get", e.getApp},
		{"app_list", e.listApps},
		{"app_update", e.updateApp},
		{"app_resource_list", e.listAppResources},
		{"route_create", e.createRoute},
		{"route_get", e.getRoute},
		{"route_list", e.listRoutes},
		{"route_delete", e.deleteRoute},
		{"artifact_create", e.createArtifact},
		{"release_create", e.createRelease},
		{"artifact_list", e.listArtifacts},
		{"release_list", e.listReleases},
		{"app_release_set", e.setAppRelease},
		{"app_release_get", e.getAppRelease},
		{"formation_put", e.putFormation},
		{"formation_get", e.getFormation},
		{"formation_list", e.listFormations},
		{"formation_delete", e.deleteFormation},
		{"job_run", e.runJob},
		{"job_list", e.listJobs},
		{"job_update", e.updateJob},
		{"job_log", e.getJobLog},
		{"job_delete", e.deleteJob},
		{"provider_create", e.createProvider},
		{"provider_get", e.getProvider},
		{"provider_list", e.listProviders},
		{"provider_resource_create", e.createProviderResource},
		{"provider_resource_get", e.getProviderResource},
		{"provider_resource_list", e.listProviderResources},
		{"app_delete", e.deleteApp},
	}

	// TODO: GET /apps/:app_id/jobs/:job_id/log (event-stream)

	res := make(map[string]string)
	for _, ex := range examples {
		ex.f()
		res[ex.name] = requestMarkdown(getRequests()[0])
	}

	var out io.Writer
	if len(os.Args) > 1 {
		out, err = os.Create(os.Args[1])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		out = os.Stdout
	}
	data, err := json.MarshalIndent(res, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	_, err = out.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *generator) listenAndServe(l *log.Logger) {
	l.Printf("Starting mock provider server on port %s\n", e.conf.ourPort)
	http.HandleFunc("/providers/", func(w http.ResponseWriter, r *http.Request) {
		l.Printf("%s %s\n", r.Method, r.URL)
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()
		l.Printf("\t%s\n", body)

		resource := &resource.Resource{
			Env: map[string]string{
				"some": "data",
			},
		}
		err := json.NewEncoder(w).Encode(resource)
		if err != nil {
			l.Println(err)
			w.WriteHeader(500)
			return
		}
	})

	http.ListenAndServe(":"+e.conf.ourPort, nil)
}

func generatePublicKey() (string, error) {
	key := `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDPI19fkFmPNg3MGqJorFTbetPJjxlhLDUJFALYe5DyqW0lAnb2R7XvXzj+kRX9LkwOeQjf6nM4bcXbd/H3YPlMDc9JfDuSGlwvo0X8KUQ6PopgyfQ15GA+8YDgwYcBJowIXqAc52GVNnBUeoZzBKvNnsVjAw6KkTPS0aZ6KBZadtYx+Y1fJJBoygh/gtPZ/MQry3XQRvbKPa0iU34Wcx8pXx5QVFLHvyORczQlEVyq5qa5DT86CRR/wC4yH32hkNGalGXY7sZg0j4EY4AeD2yCcmsp7hTt4Ql4gRp3r04ye4DZ7epdXW2tp2vJ3IVn+l6BSNooBIfoD7ZdkUVce51z some-comment`
	return key, nil
}

func (e *generator) createKey() {
	pubKey, err := generatePublicKey()
	key, err := e.client.CreateKey(pubKey)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["key"] = key.ID
}

func (e *generator) getKey() {
	e.client.GetKey(e.resourceIds["key"])
}

func (e *generator) listKeys() {
	e.client.KeyList()
}

func (e *generator) deleteKey() {
	e.client.DeleteKey(e.resourceIds["key"])
}

func (e *generator) createApp() {
	t := time.Now().UnixNano()
	app := &ct.App{Name: fmt.Sprintf("my-app-%d", t)}
	err := e.client.CreateApp(app)
	if err == nil {
		e.resourceIds["app"] = app.ID
	}
}

func (e *generator) getApp() {
	e.client.GetApp(e.resourceIds["app"])
}

func (e *generator) listApps() {
	e.client.AppList()
}

func (e *generator) updateApp() {
	app := &ct.App{
		ID: e.resourceIds["app"],
		Meta: map[string]string{
			"bread": "with hemp",
		},
	}
	e.client.UpdateApp(app)
}

func (e *generator) listAppResources() {
	e.client.AppResourceList(e.resourceIds["app"])
}

func (e *generator) createRoute() {
	config := json.RawMessage(`{
    "domain": "http://example.com"
  }`)
	route := &router.Route{
		Type:   "http",
		Config: &config,
	}
	err := e.client.CreateRoute(e.resourceIds["app"], route)
	if err == nil {
		e.resourceIds["route"] = route.ID
	}
}

func (e *generator) getRoute() {
	e.client.GetRoute(e.resourceIds["app"], e.resourceIds["route"])
}

func (e *generator) listRoutes() {
	e.client.RouteList(e.resourceIds["app"])
}

func (e *generator) deleteRoute() {
	e.client.DeleteRoute(e.resourceIds["app"], e.resourceIds["route"])
}

func (e *generator) deleteApp() {
	e.client.DeleteApp(e.resourceIds["app"])
}

func (e *generator) createArtifact() {
	artifact := &ct.Artifact{
		Type: "docker",
		URI:  "example://uri",
	}
	err := e.client.CreateArtifact(artifact)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["artifact"] = artifact.ID
}

func (e *generator) listArtifacts() {
	e.client.ArtifactList()
}

func (e *generator) createRelease() {
	release := &ct.Release{
		ArtifactID: e.resourceIds["artifact"],
		Env: map[string]string{
			"some": "info",
		},
		Processes: map[string]ct.ProcessType{
			"foo": ct.ProcessType{
				Cmd: []string{"ls", "-l"},
				Env: map[string]string{
					"BAR": "baz",
				},
			},
		},
	}
	err := e.client.CreateRelease(release)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["release"] = release.ID
}

func (e *generator) listReleases() {
	e.client.ReleaseList()
}

func (e *generator) getAppRelease() {
	e.client.GetAppRelease(e.resourceIds["app"])
}

func (e *generator) setAppRelease() {
	e.client.SetAppRelease(e.resourceIds["app"], e.resourceIds["release"])
}

func (e *generator) putFormation() {
	formation := &ct.Formation{
		AppID:     e.resourceIds["app"],
		ReleaseID: e.resourceIds["release"],
		Processes: map[string]int{
			"foo": 1,
		},
	}
	e.client.PutFormation(formation)
}

func (e *generator) getFormation() {
	e.client.GetFormation(e.resourceIds["app"], e.resourceIds["release"])
}

func (e *generator) listFormations() {
	e.client.FormationList(e.resourceIds["app"])
}

func (e *generator) deleteFormation() {
	e.client.DeleteFormation(e.resourceIds["app"], e.resourceIds["release"])
}

func (e *generator) runJob() {
	new_job := &ct.NewJob{
		ReleaseID: e.resourceIds["release"],
		Env: map[string]string{
			"BODY": "Hello!",
		},
		Cmd: []string{"echo", "$BODY"},
	}
	job, err := e.client.RunJobDetached(e.resourceIds["app"], new_job)
	if err == nil {
		e.resourceIds["job"] = job.ID
	}
}

func (e *generator) listJobs() {
	e.client.JobList(e.resourceIds["app"])
}

func (e *generator) updateJob() {
	job := &ct.Job{
		ID:        e.resourceIds["job"],
		AppID:     e.resourceIds["app"],
		ReleaseID: e.resourceIds["release"],
		State:     "down",
	}
	e.client.PutJob(job)
}

func (e *generator) getJobLog() {
	res, err := e.client.GetJobLog(e.resourceIds["app"], e.resourceIds["job"], false)
	if err == nil {
		io.Copy(ioutil.Discard, res)
	}
}

func (e *generator) deleteJob() {
	e.client.DeleteJob(e.resourceIds["app"], e.resourceIds["job"])
}

func (e *generator) createProvider() {
	t := time.Now().UnixNano()
	provider := &ct.Provider{
		Name: fmt.Sprintf("example-provider-%d", t),
		URL:  fmt.Sprintf("discoverd+http://example-provider-%d/providers/%d", t, t),
	}
	err := e.client.CreateProvider(provider)
	if err != nil {
		log.Fatal(err)
	}
	err = discoverd.Register(provider.Name, net.JoinHostPort(e.conf.ourAddr, e.conf.ourPort))
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["provider"] = provider.ID
}

func (e *generator) getProvider() {
	e.client.GetProvider(e.resourceIds["provider"])
}

func (e *generator) listProviders() {
	e.client.ProviderList()
}

func (e *generator) createProviderResource() {
	resourceReq := &ct.ResourceReq{
		ProviderID: e.resourceIds["provider"],
	}
	resource, err := e.client.ProvisionResource(resourceReq)
	if err != nil {
		log.Fatal(err)
	}
	e.resourceIds["provider_resource"] = resource.ID
}

func (e *generator) getProviderResource() {
	providerID := e.resourceIds["provider"]
	resourceID := e.resourceIds["provider_resource"]
	e.client.GetResource(providerID, resourceID)
}

func (e *generator) listProviderResources() {
	e.client.ResourceList(e.resourceIds["provider"])
}
