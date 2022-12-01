import { queryResolverFromEnv } from "./resolver";

test("dataapi queryResolverFromEnv", () => {
  queryResolverFromEnv({
    DB_CONN_TYPE: "dataapi",
    DB_REGION: "eu-west-2",
    DB_RESOURCE_ARN:
      "arn:aws:rds:eu-west-2:124567901011:cluster:dev-keel-sharedstagingdb",
    DB_SECRET_ARN:
      "arn:aws:rds:eu-west-2:124567901011:cluster:dev-keel-sharedstagingdb",
    DB_NAME: "env_2H5IwJ1PXKGtwBvkdProxySUWlt",
  });
});
