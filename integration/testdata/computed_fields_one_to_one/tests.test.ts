import { test, expect, beforeEach } from "vitest";
import { models, resetDatabase , actions} from "@teamkeel/testing";

beforeEach(resetDatabase);

test("computed fields - one to one", async () => {
  const taxProfile = await actions.createTaxProfile({ taxNumber: 1234567890 });
  const companyProfile = await actions.createCompanyProfile({ employeeCount: 100, taxProfile: { id: taxProfile.id } });
  const company = await actions.createCompany({  companyProfile: { id: companyProfile.id }, retrenchments: 8 });
  expect(company.activeEmployees).toEqual(92);
  expect(company.taxNumber).toEqual(1234567890);
});
