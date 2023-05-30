import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("create operation - set optional model field to value", async () => {
  const person = await actions.createPerson({
    nickName: {
      value: "Biggy",
    },
  });

  expect(person.nickName).toEqual("Biggy");
});

test("create operation - set optional relationship field to value", async () => {
  const company = await models.company.create({ tradingAs: "Big Company" });

  const person = await actions.createPerson({
    company: {
      value: {
        id: company.id,
      },
    },
  });

  expect(person.companyId).toEqual(company.id);
});

test("create operation - set optional model field to null", async () => {
  const person = await actions.createPerson({
    nickName: {
      isNull: true,
    },
  });

  expect(person.nickName).toBeNull();
});

test("create operation - set optional relationship field to null", async () => {
  const person = await actions.createPerson({
    company: {
      isNull: true,
    },
  });

  expect(person.companyId).toBeNull();
});

test("create operation - set optional model field to override default", async () => {
  let person = await actions.createPerson();

  expect(person.withDefault).toEqual("default value");

  person = await actions.createPerson({ withDefault: { isNull: true } });

  expect(person.withDefault).toBeNull();
});

test("create operation - create nested model optional field to value", async () => {
  const person = await actions.createPersonAndCompany({
    company: {
      value: {
        tradingAs: {
          value: "Big Company",
        },
      },
    },
  });

  const company = await models.company.findOne({ id: person.companyId! });

  expect(company?.tradingAs).toEqual("Big Company");
});

test("create operation - create nested model optional field to null", async () => {
  const person = await actions.createPersonAndCompany({
    company: {
      value: {
        tradingAs: {
          isNull: true,
        },
      },
    },
  });

  const company = await models.company.findOne({ id: person.companyId! });

  expect(company?.tradingAs).toBeNull();
});

test("create operation - create nested model field to null overriding default", async () => {
  let person = await actions.createPersonAndCompany({
    company: {
      value: {
        tradingAs: {
          value: "Big Company",
        },
      },
    },
  });

  let company = await models.company.findOne({ id: person.companyId! });
  expect(company?.withDefault).toEqual("default value");

  person = await actions.createPersonAndCompany({
    company: {
      value: {
        withDefault: {
          isNull: true,
        },
      },
    },
  });

  company = await models.company.findOne({ id: person.companyId! });

  expect(company?.withDefault).toBeNull();
});

test("update operation - set optional model field to value", async () => {
  const { id } = await models.person.create({ nickName: null });

  const person = await actions.updatePerson({
    where: {
      id: id,
    },
    values: {
      nickName: {
        value: "Biggy",
      },
    },
  });

  expect(person.nickName).toEqual("Biggy");
});

test("update operation - set optional model field to null", async () => {
  const { id } = await models.person.create({ nickName: "Biggy" });

  const person = await actions.updatePerson({
    where: {
      id: id,
    },
    values: {
      nickName: {
        isNull: true,
      },
    },
  });

  expect(person.nickName).toBeNull();
});

test("update operation - set optional relationship to value", async () => {
  const company = await models.company.create({ tradingAs: "Big Company" });
  const { id } = await models.person.create({ companyId: null });

  const person = await actions.updatePerson({
    where: {
      id: id,
    },
    values: {
      company: {
        value: {
          id: company.id,
        },
      },
    },
  });

  expect(person.companyId).toEqual(company.id);
});

test("update operation - set optional relationship to null", async () => {
  const company = await models.company.create({ tradingAs: "Big Company" });
  const { id } = await models.person.create({ companyId: company.id });

  const person = await actions.updatePerson({
    where: {
      id: id,
    },
    values: {
      company: {
        isNull: true,
      },
    },
  });

  expect(person.companyId).toBeNull();
});

test("list operation - filter by isNull on optional model field", async () => {
  const { id: id1 } = await models.person.create({ nickName: null });
  const { id: id2 } = await models.person.create({ nickName: "Bob" });

  const noNickName = await actions.listPerson({
    where: {
      nickName: {
        isNull: true,
      },
    },
  });

  expect(noNickName.results.length).toEqual(1);
  expect(noNickName.results[0].id).toEqual(id1);

  const withNickName = await actions.listPerson({
    where: {
      nickName: {
        isNull: false,
      },
    },
  });

  expect(withNickName.results.length).toEqual(1);
  expect(withNickName.results[0].id).toEqual(id2);
});

test("list operation - filter by isNull on nested optional model field", async () => {
  const company1 = await models.company.create({ tradingAs: null });
  const company2 = await models.company.create({ tradingAs: "Big Company" });

  const { id: id1 } = await models.person.create({ companyId: company1.id });
  const { id: id2 } = await models.person.create({ companyId: company2.id });

  const noNickName = await actions.listPerson({
    where: {
      companyTradingAs: {
        isNull: true,
      },
    },
  });

  expect(noNickName.results.length).toEqual(1);
  expect(noNickName.results[0].id).toEqual(id1);

  const withNickName = await actions.listPerson({
    where: {
      companyTradingAs: {
        isNull: false,
      },
    },
  });

  expect(withNickName.results.length).toEqual(1);
  expect(withNickName.results[0].id).toEqual(id2);
});
