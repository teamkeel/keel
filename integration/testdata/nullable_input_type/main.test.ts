import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";
import { Person, Company } from "@teamkeel/sdk";

beforeEach(resetDatabase);

// test("create operation - set optional model field to value", async () => {
//     let person =  await actions.createPerson({ 
//         nickName: { 
//             value: "Biggy" 
//         } 
//     });

//     expect(person.nickName).toEqual("Biggy");
// });

// test("create operation - set optional relationship field to value", async () => {
//     let company = await models.company.create({ tradingAs: "Big Company"})

//     let person =  await actions.createPerson({ 
//         company: { 
//             value: {
//                 id: company.id
//             }
//         }
//     });

//     expect(person.companyId).toEqual(company.id);
// });

// test("create operation - set optional model field to null", async () => {
//     let person =  await actions.createPerson({ 
//         nickName: { 
//             isNull: true
//         } 
//     });

//     expect(person.nickName).toBeNull();
// });

// test("create operation - set optional relationship field to null", async () => {
//     let person =  await actions.createPerson({ 
//         company: { 
//             isNull: true
//         }
//     });

//     expect(person.companyId).toBeNull();
// });


test("update operation - set optional model field to value", async () => {
    let { id } = await models.person.create({ nickName: null });
    
    let person = await actions.updatePerson({ 
        where: {
            id: id
        },
        values: {
            nickName: { 
                value: "Biggy" 
            } 
        }
    });

    expect(person.nickName).toEqual("Biggy");
});

// test("update operation - set optional relationship field to value", async () => {
//     let company = await models.company.create({ tradingAs: "Big Company"})
//     let { id } = await models.person.create({ companyId: null });

//     let person =  await actions.createPerson({ 
//         company: { 
//             value: {
//                 id: company.id 
//             }
//         }
//     });

//     expect(person.companyId).toEqual(company.id);
// });

// test("update operation - set optional model field to null", async () => {
//     let { id } = await models.person.create({ nickName: "Biggy" });
    
//     let person =  await actions.createPerson({ 
//         nickName: { 
//             isNull: true
//         } 
//     });

//     expect(person.nickName).toBeNull();
// });

test("create operation - create nested model field to value", async () => {
    let person =  await actions.createPersonAndCompany({ 
        company: { 
            value: {
                tradingAs: { 
                    value: "Big Company"
                }
            }
        }
    });

    let company = await models.company.findOne({ id: person.companyId! });

    expect(company?.tradingAs).toEqual("Big Company");
});