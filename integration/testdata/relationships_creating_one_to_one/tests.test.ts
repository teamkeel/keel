import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

// test("create relationships - one to one", async () => {

//     const country = await actions.createCountry({
//         name: "Australia",
//         cities: [
//             { name: "Canberra" },
//             { name: "Sydney" },
//             { name: "Melbourne" },
//             { name: "Brisbane" },
//         ]
//     })

//     const cities = await models.city.findMany({ where: { countryId: country.id }});
//     expect(cities).toHaveLength(4);

//     const capital = await models.city.findOne({ where: { name: "Canberra" }});
//     expect(capital).not.toBeNull();


// });

test("create relationships - one to one", async () => {

    const country = await actions.createCountry({
        name: "Australia",
        capital: { 
            name: "Canberra" 
        }
    })

    const cities = await models.city.findMany({ where: { countryId: country.id }});
    expect(cities).toHaveLength(1);
    expect(cities[0].name).toEqual("Canberra");
    expect(cities[0].capitalOf.id).toEqual(country.id);


    const capital = await models.city.findOne({ where: { name: "Canberra" }});
    expect(capital).not.toBeNull();

    
});

