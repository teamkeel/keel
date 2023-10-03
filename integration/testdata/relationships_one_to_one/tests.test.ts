import { models, actions, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);

test("create nested 1:1", async () => {
  const company = await actions.createCompany({
    name: "Bobs Shop",
    companyProfile: {
      employeeCount: 4,
      taxProfile: null,
    },
  });
  expect(company.companyProfileId).not.toBeNull();

  const companyProfile = await models.companyProfile.findOne({
    id: company.companyProfileId,
  });
  expect(companyProfile!.employeeCount).toEqual(4);
  expect(companyProfile!.taxProfileId).toBeNull();
});

test("create nested 1:1 multiple layered", async () => {
  const company = await actions.createCompany({
    name: "Bobs Shop",
    companyProfile: {
      employeeCount: 4,
      taxProfile: {
        taxNumber: "DSJK7722S",
      },
    },
  });
  expect(company.companyProfileId).not.toBeNull();

  const companyProfile = await models.companyProfile.findOne({
    id: company.companyProfileId,
  });
  expect(companyProfile!.employeeCount).toEqual(4);
  expect(companyProfile!.taxProfileId).not.toBeNull();

  const taxProfile = await models.taxProfile.findOne({
    id: companyProfile!.taxProfileId!,
  });
  expect(taxProfile!.taxNumber).toEqual("DSJK7722S");
});

test("get nested 1:1 by implicit inputs (forward traversal)", async () => {
  const company1 = await actions.createCompany({
    name: "Bobs Shop",
    companyProfile: {
      employeeCount: 4,
      taxProfile: {
        taxNumber: "DSJK7722S",
      },
    },
  });

  const company2 = await actions.createCompany({
    name: "Rudolfs Shop",
    companyProfile: {
      employeeCount: 4,
      taxProfile: {
        taxNumber: "11FFWWEF",
      },
    },
  });

  const company = await actions.getCompanyByTaxNumber({
    companyProfileTaxProfileTaxNumber: "11FFWWEF",
  });

  expect(company!.id).toEqual(company2.id);
});

test("list nested 1:1 by expressions (forward traversal)", async () => {
  await actions.createCompany({
    name: "Bobs Shop",
    companyProfile: {
      employeeCount: 4,
      taxProfile: {
        taxNumber: "DSJK7722S",
      },
    },
  });

  const largeCompany = await actions.createCompany({
    name: "Dylans Shop",
    companyProfile: {
      employeeCount: 150,
      taxProfile: {
        taxNumber: "11FFWWEF",
      },
    },
  });

  await actions.createCompany({
    name: "Andys Shop",
    companyProfile: {
      employeeCount: 118,
    },
  });

  const companies = await actions.largeCompaniesRegistered();
  expect(companies.pageInfo.count).toEqual(1);
  expect(companies.results[0].id).toEqual(largeCompany.id);
  expect(companies.results[0].name).toEqual("Dylans Shop");
});

test("list nested 1:1 by implicit inputs (backwards traversal)", async () => {
  const company = await actions.createCompany({
    name: "Bobs Shop",
    companyProfile: {
      employeeCount: 4,
      taxProfile: {
        taxNumber: "DSJK7722S",
      },
    },
  });
  expect(company.companyProfileId).not.toBeNull();

  const companyProfiles = await actions.findCompanyProfile({
    where: {
      company: {
        id: {
          equals: company.id,
        },
      },
    },
  });
  expect(companyProfiles.pageInfo.count).toEqual(1);
  expect(companyProfiles.results[0].id).toEqual(company.companyProfileId);
  expect(companyProfiles.results[0].employeeCount).toEqual(4);
  expect(companyProfiles.results[0].taxProfileId).not.toBeNull();

  const taxProfiles = await actions.findTaxProfile({
    where: {
      companyProfile: {
        company: {
          id: {
            equals: company.id,
          },
        },
      },
    },
  });
  expect(taxProfiles.pageInfo.count).toEqual(1);
  expect(taxProfiles.results[0].id).toEqual(
    companyProfiles.results[0].taxProfileId
  );
  expect(taxProfiles.results[0].taxNumber).toEqual("DSJK7722S");
});

test("create nested 1:1 with duplicated nested unique value", async () => {
  await actions.createCompany({
    name: "Bobs Shop",
    companyProfile: {
      employeeCount: 4,
      taxProfile: {
        taxNumber: "DSJK7722S",
      },
    },
  });

  await expect(
    actions.createCompany({
      name: "Andys Shop",
      companyProfile: {
        employeeCount: 36,
        taxProfile: {
          taxNumber: "DSJK7722S",
        },
      },
    })
  ).toHaveError({
    code: "ERR_INVALID_INPUT",
    message: "the value for the unique field 'taxNumber' must be unique",
  });
});
