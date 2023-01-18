const { Kysely, PostgresDialect } = require("kysely");
const pg = require("pg");
const { DataApiDialect } = require("kysely-data-api");
const RDSDataService = require("aws-sdk/clients/rdsdataservice");

function mustEnv(key) {
  const v = process.env[key];
  if (!v) {
    throw new Error(`expected environment variable ${key} to be set`);
  }
  return v;
}

function getDialect() {
  const dbConnType = process.env["DB_CONN_TYPE"];
  switch (dbConnType) {
    case "pg":
      return new PostgresDialect({
        pool: new pg.Pool({
          connectionString: mustEnv("DB_CONN"),
        }),
      });

    case "dataapi":
      return new DataApiDialect({
        mode: "postgres",
        driver: {
          client: new RDSDataService({
            region: mustEnv("DB_REGION"),
          }),
          database: mustEnv("DB_NAME"),
          secretArn: mustEnv("DB_SECRET_ARN"),
          resourceArn: mustEnv("DB_RESOURCE_ARN"),
        },
      });

    default:
      throw Error("unexpected DB_CONN_TYPE: " + dbConnType);
  }
}

let db = null;

function getDatabase() {
  if (db) {
    return db;
  }

  db = new Kysely({
    dialect: getDialect(),
  });

  return db;
}

module.exports.getDatabase = getDatabase;
