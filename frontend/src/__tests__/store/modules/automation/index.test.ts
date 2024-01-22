import "jest";
import { describe, expect, test } from "@jest/globals";
import { store_temp } from "../../../../store/index";
import { default as automation1 } from "../../../../../../core/configs/automations/0x001788010d7d9d3f.json"
import { default as automation2 } from "../../../../../../core/configs/automations/0xa4c13894070052fc.json"

import { Automation, Automations } from "../../../../types/automation";

test("test automations/find can load an inserted automation", () => {
  // add them to store
  var json = JSON.stringify(automation1);
  const newAutomation: Automation = JSON.parse(json);
  store_temp.commit("automations/add", newAutomation);

  // assert
  const result = store_temp.getters["automations/find"](newAutomation.id) as Automation;
  expect(result).toEqual(newAutomation);
})

test("test automations/listAll returns a list of all automations", () => {

  const testData = [automation1, automation2]
  const automations: Automations = []
  // add them to store
  testData.forEach(data => {
    var json = JSON.stringify(data);
    const newAutomation: Automation = JSON.parse(json);
    store_temp.commit("automations/add", newAutomation);

    automations.push(newAutomation)
  });

  // assert
  const result = store_temp.getters["automations/listAll"]() as Automations;

  automations.forEach((automation, idx) => {
    expect(automation).toEqual(result[idx]);
  })

});

test("test automations/initialized returns correct value ", () => {

  // assert value before initialization
  var result = store_temp.getters["automations/initialized"]() as boolean;
  expect(result).toEqual(false);

  const testData = [automation1, automation2]
  const automations: Automations = []

  // add them to store
  testData.forEach(data => {
    var json = JSON.stringify(data);
    const newAutomation: Automation = JSON.parse(json);
    automations.push(newAutomation)
  });

  store_temp.dispatch("automations/init", automations);

  // assert values after initialization
  var result = store_temp.getters["automations/initialized"]() as boolean;
  expect(result).toEqual(true);
});

test("test automations/init stores the items ", () => {

  const testData = [automation1, automation2]
  const automations: Automations = []

  // add them to store
  testData.forEach(data => {
    var json = JSON.stringify(data);
    const newAutomation: Automation = JSON.parse(json);
    automations.push(newAutomation)
  });

  store_temp.dispatch("automations/init", automations);

  // assert values 
  const items = store_temp.getters["automations/listAll"]() as Automations;
  automations.forEach((automation, idx) => {
    expect(automation).toEqual(items[idx]);
  })

  const result = store_temp.getters["automations/initialized"]() as boolean;
  expect(result).toEqual(true);
});

test("test automations/clear removes all automations from store", () => {
  const testData = [automation1, automation2]
  const automations: Automations = []

  // add them to store
  testData.forEach(data => {
    var json = JSON.stringify(data);
    const newAutomation: Automation = JSON.parse(json);
    automations.push(newAutomation)
  });

  store_temp.dispatch("automations/init", automations);

  // assert values
  store_temp.commit("automations/clear");
  const result = store_temp.getters["automations/listAll"]() as Automations;

  expect(result.length).toBe(0);
});

test("test automations/update operation", () => {

  // add to store
  var json = JSON.stringify(automation1);
  const newAutomation: Automation = JSON.parse(json);
  store_temp.commit("automations/add", newAutomation);

  // update data 
  newAutomation.friendlyname = "name changed"
  newAutomation.enabled = false

  newAutomation.triggers[0].name = "trigger name updated"
  newAutomation.triggers[0].conditions[0].value = 50
  newAutomation.triggers[0].action.operation = 2
  newAutomation.triggers[0].action.data = 12.9
  store_temp.commit("automations/update", newAutomation);

  // assert values
  const result = store_temp.getters['automations/find'](newAutomation.id) as Automation;
  expect(result).toEqual(newAutomation);
});


test("test automations/delete operation", () => {

  // add to store
  var json = JSON.stringify(automation1);
  const newAutomation: Automation = JSON.parse(json);
  store_temp.commit("automations/add", newAutomation);

  store_temp.commit("automations/delete", newAutomation.id);

  // assert values
  const result = store_temp.getters['automations/find'](newAutomation.id) as Automation;
  expect(result).toBeUndefined();
});


