// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state_test

import (
	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/workload"
	"github.com/juju/juju/workload/state"
)

var _ = gc.Suite(&unitWorkloadsSuite{})

type unitWorkloadsSuite struct {
	baseWorkloadsSuite
	id string
}

func (s *unitWorkloadsSuite) newID() (string, error) {
	s.stub.AddCall("newID")
	if err := s.stub.NextErr(); err != nil {
		return "", errors.Trace(err)
	}

	return s.id, nil
}

func (s *unitWorkloadsSuite) TestNewID(c *gc.C) {
	id, err := state.NewID()
	c.Assert(err, jc.ErrorIsNil)

	c.Check(id, jc.Satisfies, utils.IsValidUUIDString)
}

func (s *unitWorkloadsSuite) TestTrackOkay(c *gc.C) {
	s.id = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	wl := pl.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	ps = state.SetNewID(ps, s.newID)

	err := ps.Track(wl)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "newID", "Track")
	c.Check(s.persist.payloads, gc.HasLen, 1)
	s.persist.checkPayload(c, s.id, pl)
}

func (s *unitWorkloadsSuite) TestTrackInvalid(c *gc.C) {
	s.id = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	pl := s.newPayload("", "workloadA/workloadA-xyz")
	wl := pl.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	ps = state.SetNewID(ps, s.newID)

	err := ps.Track(wl)

	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *unitWorkloadsSuite) TestTrackEnsureDefinitionFailed(c *gc.C) {
	s.id = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	failure := errors.Errorf("<failed!>")
	s.stub.SetErrors(failure)
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	wl := pl.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	ps = state.SetNewID(ps, s.newID)

	err := ps.Track(wl)

	c.Check(errors.Cause(err), gc.Equals, failure)
}

func (s *unitWorkloadsSuite) TestTrackInsertFailed(c *gc.C) {
	s.id = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	failure := errors.Errorf("<failed!>")
	s.stub.SetErrors(failure)
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	wl := pl.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	ps = state.SetNewID(ps, s.newID)

	err := ps.Track(wl)

	c.Check(errors.Cause(err), gc.Equals, failure)
}

func (s *unitWorkloadsSuite) TestTrackAlreadyExists(c *gc.C) {
	s.id = "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	s.persist.setPayload(s.id, &pl)
	wl := pl.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	ps = state.SetNewID(ps, s.newID)

	err := ps.Track(wl)

	c.Check(err, jc.Satisfies, errors.IsNotValid)
}

func (s *unitWorkloadsSuite) TestSetStatusOkay(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	s.persist.setPayload(id, &pl)

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}

	err := ps.SetStatus(id, workload.StateRunning)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "SetStatus")
	current := s.persist.payloads[id]
	c.Check(current.Status, jc.DeepEquals, workload.StateRunning)
}

func (s *unitWorkloadsSuite) TestSetStatusFailed(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	failure := errors.Errorf("<failed!>")
	s.stub.SetErrors(failure)
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	s.persist.setPayload(id, &pl)

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	err := ps.SetStatus(id, workload.StateRunning)

	c.Check(errors.Cause(err), gc.Equals, failure)
}

func (s *unitWorkloadsSuite) TestSetStatusMissing(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	err := ps.SetStatus(id, workload.StateRunning)

	c.Check(err, jc.Satisfies, errors.IsNotFound)
}

func (s *unitWorkloadsSuite) TestListOkay(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	otherID := "f47ac10b-58cc-4372-a567-0e02b2c3d480"
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	other := s.newPayload("docker", "workloadB/workloadB-abc")
	s.persist.setPayload(id, &pl)
	s.persist.setPayload(otherID, &other)
	wl := pl.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	results, err := ps.List(id)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "List")
	c.Check(results, jc.DeepEquals, []workload.Result{{
		ID:       id,
		Workload: &wl,
	}})
}

func (s *unitWorkloadsSuite) TestListAll(c *gc.C) {
	id1 := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	id2 := "f47ac10b-58cc-4372-a567-0e02b2c3d480"
	pl1 := s.newPayload("docker", "workloadA/workloadA-xyz")
	pl2 := s.newPayload("docker", "workloadB/workloadB-abc")
	s.persist.setPayload(id1, &pl1)
	s.persist.setPayload(id2, &pl2)
	wl1 := pl1.AsWorkload()
	wl2 := pl2.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	results, err := ps.List()
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "ListAll", "LookUp", "LookUp")
	c.Assert(results, gc.HasLen, 2)
	if results[0].Workload.Name == "workloadA" {
		c.Check(results[0].Workload, jc.DeepEquals, &wl1)
		c.Check(results[1].Workload, jc.DeepEquals, &wl2)
	} else {
		c.Check(results[0].Workload, jc.DeepEquals, &wl2)
		c.Check(results[1].Workload, jc.DeepEquals, &wl1)
	}
}

func (s *unitWorkloadsSuite) TestListFailed(c *gc.C) {
	failure := errors.Errorf("<failed!>")
	s.stub.SetErrors(failure)

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	_, err := ps.List()

	s.stub.CheckCallNames(c, "ListAll")
	c.Check(errors.Cause(err), gc.Equals, failure)
}

func (s *unitWorkloadsSuite) TestListMissing(c *gc.C) {
	missingID := "f47ac10b-58cc-4372-a567-0e02b2c3d480"
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	s.persist.setPayload(id, &pl)
	wl := pl.AsWorkload()

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	results, err := ps.List(id, missingID)
	c.Assert(err, jc.ErrorIsNil)

	c.Assert(results, gc.HasLen, 2)
	c.Check(results[1].Error, gc.NotNil)
	results[1].Error = nil
	c.Check(results, jc.DeepEquals, []workload.Result{{
		ID:       id,
		Workload: &wl,
	}, {
		ID:       missingID,
		NotFound: true,
	}})
}

func (s *unitWorkloadsSuite) TestUntrackOkay(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	pl := s.newPayload("docker", "workloadA/workloadA-xyz")
	s.persist.setPayload(id, &pl)

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	err := ps.Untrack(id)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "Untrack")
	c.Check(s.persist.payloads, gc.HasLen, 0)
}

func (s *unitWorkloadsSuite) TestUntrackMissing(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	err := ps.Untrack(id)
	c.Assert(err, jc.ErrorIsNil)

	s.stub.CheckCallNames(c, "Untrack")
	c.Check(s.persist.payloads, gc.HasLen, 0)
}

func (s *unitWorkloadsSuite) TestUntrackFailed(c *gc.C) {
	id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
	failure := errors.Errorf("<failed!>")
	s.stub.SetErrors(failure)

	ps := state.UnitWorkloads{
		Persist: s.persist,
		Unit:    "a-service/0",
	}
	err := ps.Untrack(id)

	s.stub.CheckCallNames(c, "Untrack")
	c.Check(errors.Cause(err), gc.Equals, failure)
}
