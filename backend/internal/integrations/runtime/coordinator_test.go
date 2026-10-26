package runtime

import (
	"net/http"
	"testing"
)

func TestCoordinatorFencesLateResponse(t *testing.T) {
	for _, action := range []string{"disable", "delete"} {
		t.Run(action, func(t *testing.T) {
			started := make(chan struct{})
			release := make(chan struct{})
			s, _ := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path == "/v1/schedules/synthetic" {
					close(started)
					<-release
				}
				return standardProvider(r)
			}))
			i := configuredFixture(t, s)
			revision := i.Resources[0].Revision
			op, e := s.Refresh(i.ID, i.ETag)
			if e != nil {
				t.Fatal(e)
			}
			queued := <-s.queue
			finished := make(chan struct{})
			go func() { s.execute(queued); close(finished) }()
			<-started
			i, e = s.Instance(i.ID)
			if e != nil {
				t.Fatal(e)
			}
			if action == "disable" {
				_, e = s.UpdateInstance(i.ID, i.Name, false, i.ETag)
			} else {
				e = s.DeleteInstance(i.ID, i.ETag)
			}
			if e != nil {
				t.Fatal(e)
			}
			close(release)
			<-finished
			after, e := s.Instance(i.ID)
			if action == "delete" {
				if e == nil {
					t.Fatal("deleted instance resurrected")
				}
				return
			}
			if e != nil || after.Resources[0].Revision != revision || after.Enabled {
				t.Fatal("late result published", op, after, e)
			}
		})
	}
}
