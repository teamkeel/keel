import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create relationships - one to one", async () => {
  const president = await actions.createPresident({ 
    name: "President Keelinson",
    member: {
      name: "Keelinson",
      party: { id: "keeldom"}
    },
    nation: { 
      name: "Keeldom" 
    },
    party: { 
      id: "keeldom",
      name: "Backends for All",
      members: [
        { 
          name: "Keelson",
          party: { id: "keeldom" }
        },
        { 
          name: "Keeler",
          party: { id: "keeldom" }
        },
        { 
          name: "Keeliner",
          party: { id: "keeldom" }
        },
      ]
    }
  });

  const nations = await models.nation.findMany();
  expect(nations).toHaveLength(1);
  const nation = nations[0];

  const parties = await models.party.findMany();
  expect(parties).toHaveLength(1);
  const party = parties[0];

  const member = await models.member.findOne({id: president.memberId});
expect(member?.name).toEqual("Keelinson")
expect(member?.partyId).toEqual(party.id)

  expect(president.name).toEqual("President Keelinson");
  expect(president.partyId).toEqual(party.id);
  expect(nation.name).toEqual("Keeldom");
  expect(nation.presidentId).toEqual(president.id);
  expect(party.name).toEqual("Backends for All");

  const members = await models.member.findMany();
  expect(members).toHaveLength(4);

  for(var m of members) {
    expect(m.partyId).toEqual(party.id)
  }
});
