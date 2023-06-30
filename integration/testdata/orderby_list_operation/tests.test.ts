import { test, expect, beforeEach, beforeAll } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { after } from "node:test";



//beforeEach(resetDatabase);

beforeAll(async () => {
    // const teamSA = await models.team.create({ name: "South Africa" });
    // const teamUK = await models.team.create({ name: "United Kingdom" });
    // const teamAus = await models.team.create({ name: "Australia", disqualified: true });
    // await models.contestant.create({ name: "Donald", gold: 2, silver: 4, bronze: 5, teamId: teamUK.id });
    // await models.contestant.create({ name: "Bongani", gold: 4, silver: 5, bronze: 7, teamId: teamSA.id });
    // await models.contestant.create({ name: "John", gold: 4, silver: 1, bronze: 3, teamId: teamUK.id });
    // await models.contestant.create({ name: "Stoffel", gold: 4, silver: 5, bronze: 10, teamId: teamSA.id });
    // await models.contestant.create({ name: "Mary", gold: 7, silver: 1, bronze: 3, teamId: teamUK.id });
    // await models.contestant.create({ name: "Johannes", disqualified: true, gold: 3, silver: 1, bronze: 3, teamId: teamSA.id });
    // await models.contestant.create({ name: "Addison", gold: 6, silver: 6, bronze: 6, teamId: teamAus.id })
    
    await models.contestant.create({ name: "Donald", gold: 2, silver: 4, bronze: 5 }); // 5th
    await models.contestant.create({ name: "Bongani", gold: 4, silver: 5, bronze: 7 }); // 3rd
    await models.contestant.create({ name: "John", gold: 4, silver: 1, bronze: 3 }); // 4th
    await models.contestant.create({ name: "Stoffel", gold: 4, silver: 5, bronze: 10 }); // 2nd
    await models.contestant.create({ name: "Mary", gold: 7, silver: 1, bronze: 3 }); // 1st
    await models.contestant.create({ name: "Johannes", disqualified: true, gold: 3, silver: 1, bronze: 3 });
   // await models.contestant.create({ name: "Addison", gold: 6, silver: 6, bronze: 6 });;

})

// test("orderby", async () => {
//     const winners = await actions.listRankings();

//     expect(winners.pageInfo.count).toEqual(5);
//     expect(winners.results[0].name).toEqual("Mary");
//     expect(winners.results[1].name).toEqual("Stoffel");
//     expect(winners.results[2].name).toEqual("Bongani");
//     expect(winners.results[3].name).toEqual("John");
//     expect(winners.results[4].name).toEqual("Donald");
// });

// test("orderby with implicit filter", async () => {
//     const winners = await actions.listRankings({ 
//         where: {
//             team: { name: { equals: "South Africa" }}
//         }
//     });

//     expect(winners.pageInfo.count).toEqual(2);
//     expect(winners.results[0].name).toEqual("Stoffel");
//     expect(winners.results[1].name).toEqual("Bongani");
// });

// test("orderby top 3", async () => {
//     const winners = await actions.listRankings({ 
//         first: 3
//     });

//     expect(winners.pageInfo.count).toEqual(3);
//     expect(winners.results[0].name).toEqual("Mary");
//     expect(winners.results[1].name).toEqual("Stoffel");
//     expect(winners.results[2].name).toEqual("Bongani");
// });

test("orderby fourth place", async () => {
    const winners = await actions.listRankings({ 
        
    });
    //console.log(winners);
   // const cursor = winners.pageInfo.endCursor;
  // console.log(winners);
   const cursor = winners.results[2].id;
   // console.log(cursor);
    const fourth = await actions.listRankings({ 
        //first: 1,
        after: cursor
    });
    console.log(fourth.results);
    expect(fourth.pageInfo.count).toEqual(2);
    expect(fourth.results[0].name).toEqual("John");
    
});