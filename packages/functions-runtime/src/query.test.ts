import Query from "./query";
import Logger from "./logger";
import { Input } from "./types";
import { PgQueryResolver, QueryResult } from "./db/resolver";
import { rawSql } from "./db/query";

const connectionString = `postgresql://postgres:postgres@localhost:5432/functions-runtime`;
const queryResolver = new PgQueryResolver({ connectionString });
// data api:
// const queryResolver = new AwsRdsDataClientQueryResolver({
//   dbClusterResourceArn: "",
//   dbCredentialsSecretArn: "",
//   dbName: "",
//   region: "eu-west-2"
// });

async function runInitialSql(sql: string): Promise<QueryResult> {
  // don't do this outside of tests, this is vulnerable to sql injetions
  return queryResolver.runQuery([rawSql(sql)]);
}

test("select", async () => {
  interface Person {
    id: string;
    name: string;
    married: boolean;
    favouriteNumber: number;
    date: Date;
  }

  const prepareTestSql = `
    DROP TABLE IF EXISTS person;

    CREATE TABLE person(
        id               text PRIMARY KEY,
        name             text,
        married          boolean,
        favourite_number integer,
        date             timestamp
    );

    INSERT INTO person (id, name, married, favourite_number, date)
    VALUES ('6cba2acc-a06b-4f4d-8671-bd87a5473ed9', 'Jane Doe', false, 0, '2013-03-21 09:10:59.897666Z');
    INSERT INTO person (id, name, married, favourite_number, date)
    VALUES ('3469bd00-a51c-4efb-b045-4eda9823f590', 'Keel Keelson', true, 10, '2013-03-21 09:10:59.897667Z');
    INSERT INTO person (id, name, married, favourite_number, date)
    VALUES ('1f5ed3cb-1410-48cd-b499-df17d9a7906c', 'Keel Keelgrandson', false, 10, '2013-03-21 09:10:59.897667Z');
    INSERT INTO person (id, name, married, favourite_number, date)
    VALUES ('bc3033c2-f331-463a-a335-5aa1a05f990c', 'Agent Smith', false, 1, '2013-03-21 09:10:59.897668Z');
    INSERT INTO person (id, name, married, favourite_number, date)
    VALUES ('8d53f4d9-0139-48a1-b3fa-0333030c023b', null, null, null, null);
  `;
  const janeDoe = {
    id: "6cba2acc-a06b-4f4d-8671-bd87a5473ed9",
    name: "Jane Doe",
    married: false,
    favouriteNumber: 0,
    date: new Date("2013-03-21 09:10:59.897"),
  };
  const keelKeelson = {
    id: "3469bd00-a51c-4efb-b045-4eda9823f590",
    name: "Keel Keelson",
    married: true,
    favouriteNumber: 10,
    date: new Date("2013-03-21 09:10:59.897"),
  };
  const keelKeelgrandson = {
    id: "1f5ed3cb-1410-48cd-b499-df17d9a7906c",
    name: "Keel Keelgrandson",
    married: false,
    favouriteNumber: 10,
    date: new Date("2013-03-21 09:10:59.897"),
  };
  const agentSmith = {
    id: "bc3033c2-f331-463a-a335-5aa1a05f990c",
    name: "Agent Smith",
    married: false,
    favouriteNumber: 1,
    date: new Date("2013-03-21 09:10:59.897"),
  };
  const nullPerson = {
    id: "8d53f4d9-0139-48a1-b3fa-0333030c023b",
    name: null,
    married: null,
    favouriteNumber: null,
    date: null,
  };

  await runInitialSql(prepareTestSql);

  const tableName = "person";

  const logger = new Logger();
  const query = new Query<Person>({
    tableName,
    queryResolver,
    logger,
  });

  expect(await query.all()).toEqual({
    collection: [
      janeDoe,
      keelKeelson,
      keelKeelgrandson,
      agentSmith,
      nullPerson,
    ],
  });

  expect(await query.where({ id: { equals: janeDoe.id } }).all()).toEqual({
    collection: [janeDoe],
  });

  expect(await query.where({ id: { equals: janeDoe.id } }).findOne()).toEqual({
    errors: [],
    object: janeDoe,
  });

  expect(await query.where({ id: janeDoe.id }).findOne()).toEqual({
    errors: [],
    object: janeDoe,
  });

  expect(await query.where({ name: { contains: "o" } }).all()).toEqual({
    collection: [janeDoe, keelKeelson, keelKeelgrandson],
  });

  expect(await query.where({ name: { contains: "Smith" } }).findOne()).toEqual({
    errors: [],
    object: agentSmith,
  });

  expect(await query.where({ name: { startsWith: "K" } }).all()).toEqual({
    collection: [keelKeelson, keelKeelgrandson],
  });

  expect(await query.where({ name: { endsWith: "son" } }).all()).toEqual({
    collection: [keelKeelson, keelKeelgrandson],
  });

  expect(
    await query.where({ id: { oneOf: [janeDoe.id, agentSmith.id] } }).all()
  ).toEqual({ collection: [janeDoe, agentSmith] });

  expect(await query.where({ id: { notEquals: janeDoe.id } }).all()).toEqual({
    collection: [keelKeelson, keelKeelgrandson, agentSmith, nullPerson],
  });

  expect(await query.where({ married: { equals: true } }).all()).toEqual({
    collection: [keelKeelson],
  });

  expect(await query.where({ married: { equals: false } }).all()).toEqual({
    collection: [janeDoe, keelKeelgrandson, agentSmith],
  });

  expect(await query.where({ married: { notEquals: true } }).all()).toEqual({
    collection: [janeDoe, keelKeelgrandson, agentSmith, nullPerson],
  });

  expect(await query.where({ married: { equals: null } }).all()).toEqual({
    collection: [nullPerson],
  });

  expect(await query.where({ favouriteNumber: { equals: 10 } }).all()).toEqual({
    collection: [keelKeelson, keelKeelgrandson],
  });

  expect(
    await query.where({ favouriteNumber: { greaterThan: 1 } }).all()
  ).toEqual({ collection: [keelKeelson, keelKeelgrandson] });

  expect(
    await query.where({ favouriteNumber: { greaterThanOrEquals: 1 } }).all()
  ).toEqual({ collection: [keelKeelson, keelKeelgrandson, agentSmith] });

  expect(
    await query
      .where({ favouriteNumber: { lessThan: 1 } })
      .orWhere({ favouriteNumber: { greaterThan: 1 } })
      .all()
  ).toEqual({ collection: [janeDoe, keelKeelson, keelKeelgrandson] });

  expect(
    await query
      .where({ favouriteNumber: { lessThan: 1 } })
      .orWhere({ id: keelKeelson.id, favouriteNumber: { greaterThan: 1 } })
      .all()
  ).toEqual({ collection: [janeDoe, keelKeelson] });

  expect(
    await query
      .where({ favouriteNumber: { lessThan: 1 } })
      .orWhere({ id: keelKeelson.id, favouriteNumber: { greaterThan: 1 } })
      .orWhere({
        name: keelKeelgrandson.name,
        favouriteNumber: { greaterThan: 1 },
      })
      .all()
  ).toEqual({ collection: [janeDoe, keelKeelson, keelKeelgrandson] });

  expect(
    await query
      .where({ favouriteNumber: { lessThan: 1 } })
      .orWhere({ id: keelKeelson.id, favouriteNumber: { greaterThan: 1 } })
      .orWhere({
        name: keelKeelgrandson.name,
        favouriteNumber: { greaterThan: 10 },
      })
      .all()
  ).toEqual({ collection: [janeDoe, keelKeelson] });

  expect(
    await query.where({ favouriteNumber: { lessThanOrEquals: 1 } }).all()
  ).toEqual({ collection: [janeDoe, agentSmith] });

  expect(
    await query
      .where({ married: { equals: false } })
      .order({ favouriteNumber: "DESC" })
      .all()
  ).toEqual({
    collection: [keelKeelgrandson, agentSmith, janeDoe],
  });

  expect(
    await query
      .where({ married: { equals: false } })
      .order({ favouriteNumber: "ASC" })
      .all()
  ).toEqual({
    collection: [janeDoe, agentSmith, keelKeelgrandson],
  });
});

test("insert", async () => {
  interface Post {
    id: string;
    title: string;
    published: boolean;
    relevance: number;
    author_born_in: Date;
  }
  type CreatePost = Partial<Omit<Post, "id">>;

  const prepareTestSql = `
    DROP TABLE IF EXISTS post;

    CREATE TABLE post(
        id             text PRIMARY KEY,
        title          text,
        published      boolean,
        relevance      integer,
        author_born_in timestamp,
        created_at     timestamp DEFAULT now(),
        updated_at     timestamp DEFAULT now()
    );
  `;

  const tableName = "post";

  const logger = new Logger();
  const query = new Query<Post>({
    tableName,
    queryResolver,
    logger,
  });

  await runInitialSql(prepareTestSql);

  expect(await query.all()).toEqual({ collection: [] });

  let postToCreate: CreatePost = {
    title: "The Most Amazing News",
    relevance: 9000,
    published: true,
    author_born_in: new Date("2013-03-21 09:10:59.897Z"),
  };

  let queryResult = await query.create(postToCreate);

  expect(Object.keys(queryResult.object)).toEqual([
    "title",
    "relevance",
    "published",
    "author_born_in",
    "id",
  ]);
  expect(queryResult.object?.id).toBeTruthy();
  expect(queryResult.object?.title).toEqual(postToCreate.title);
  expect(queryResult.object?.relevance).toEqual(postToCreate.relevance);
  expect(queryResult.object?.published).toEqual(postToCreate.published);
  expect(queryResult.object?.author_born_in).toEqual(
    postToCreate.author_born_in
  );

  postToCreate = {
    title: null,
    relevance: null,
    published: null,
    author_born_in: null,
  };

  queryResult = await query.create(postToCreate);

  expect(Object.keys(queryResult.object)).toEqual([
    "title",
    "relevance",
    "published",
    "author_born_in",
    "id",
  ]);
  expect(queryResult.object?.id).toBeTruthy();
  expect(queryResult.object?.title).toEqual(postToCreate.title);
  expect(queryResult.object?.relevance).toEqual(postToCreate.relevance);
  expect(queryResult.object?.published).toEqual(postToCreate.published);
  expect(queryResult.object?.author_born_in).toEqual(
    postToCreate.author_born_in
  );
});

test("rawSql", async () => {
  interface Food {
    id: string;
    name?: string;
    rotten?: boolean;
    //TODO add date here
    stock?: number;
  }

  const prepareTestSql = `
  DROP TABLE IF EXISTS food;

  CREATE TABLE food(
      id         text PRIMARY KEY,
      name       text,
      rotten     boolean,
      stock      integer
  );

  INSERT INTO food (id, name, rotten, stock)
  VALUES ('6cba2acc-a06b-4f4d-8671-bd87a5473ed9', 'Rotten apple', true, 10);
  INSERT INTO food (id, name, rotten, stock)
  VALUES ('6cba2acc-a06b-4f4d-8671-bd87a5473ed3', 'Gone off pineapple', true, 10);
  INSERT INTO food (id, name, rotten, stock)
  VALUES ('5a09be63-190f-4c77-a297-b4be4c023b71', 'Big Mac', false, 1);
`;

  const tableName = "food";

  const logger = new Logger();
  const query = new Query<Food>({
    tableName,
    queryResolver,
    logger,
  });

  await runInitialSql(prepareTestSql);

  const res = await query.raw(
    "SELECT CASE WHEN rotten THEN 'rotten' ELSE 'fresh' END as status, count(*) as count from food GROUP BY rotten ORDER BY 2 DESC"
  );

  interface QueryResult {
    status: string;
    count: number;
  }

  const results: QueryResult[] = res.map((r) => ({
    status: r["status"],
    count: parseInt(r["count"], 10),
  }));

  expect(results.length).toEqual(2);

  expect(results[0].count).toEqual(2);
  expect(results[0].status).toEqual("rotten");

  expect(results[1].count).toEqual(1);
  expect(results[1].status).toEqual("fresh");
});

test("delete", async () => {
  interface Animal {
    id: string;
    name: string;
  }

  const prepareTestSql = `
    DROP TABLE IF EXISTS animal;

    CREATE TABLE animal(
        id         text PRIMARY KEY,
        name       text
    );

    INSERT INTO animal(id, name) VALUES ('5a09be63-190f-4c77-a297-b4be4c023b71', 'Scooby Doo');
    INSERT INTO animal(id, name) VALUES ('66c83d78-25b9-4794-aca5-701a46bed575', 'Snoopy');
  `;

  const tableName = "animal";

  const logger = new Logger();
  const query = new Query<Animal>({
    tableName,
    queryResolver,
    logger,
  });

  await runInitialSql(prepareTestSql);

  let scoobyDoo = {
    id: "5a09be63-190f-4c77-a297-b4be4c023b71",
    name: "Scooby Doo",
  };
  let snoopy = { id: "66c83d78-25b9-4794-aca5-701a46bed575", name: "Snoopy" };

  expect(await query.all()).toEqual({ collection: [scoobyDoo, snoopy] });

  expect(await query.delete(snoopy.id)).toEqual({ success: true });

  expect(await query.delete(snoopy.id)).toEqual({ success: false });

  expect(await query.all()).toEqual({ collection: [scoobyDoo] });

  expect(await query.delete(scoobyDoo.id)).toEqual({ success: true });

  expect(await query.all()).toEqual({ collection: [] });
});

test("update", async () => {
  interface Food {
    id: string;
    name?: string;
    rotten?: boolean;
    //TODO add date here
    stock?: number;
  }
  type UpdateFood = Input<Food>;

  const prepareTestSql = `
    DROP TABLE IF EXISTS food;

    CREATE TABLE food(
        id         text PRIMARY KEY,
        name       text,
        rotten     boolean,
        stock      integer
    );

    INSERT INTO food (id, name, rotten, stock)
    VALUES ('6cba2acc-a06b-4f4d-8671-bd87a5473ed9', 'Apple', false, 10);
    INSERT INTO food (id, name, rotten, stock)
    VALUES ('414467c1-817c-4bf2-8911-b1df8e689806', 'Burger', true, 1);
  `;

  const tableName = "food";

  const logger = new Logger();
  const query = new Query<Food>({
    tableName,
    queryResolver,
    logger,
  });

  await runInitialSql(prepareTestSql);

  let apple = {
    id: "6cba2acc-a06b-4f4d-8671-bd87a5473ed9",
    name: "Apple",
    rotten: false,
    stock: 10,
  };
  let burger = {
    id: "414467c1-817c-4bf2-8911-b1df8e689806",
    name: "Burger",
    rotten: true,
    stock: 1,
  };
  let updatedApple1: UpdateFood = { ...apple, name: "Pear" };
  let updatedApple2: UpdateFood = { ...apple, id: "updated_id" };
  let updatedApple3: UpdateFood = {
    id: updatedApple2.id,
    name: "Onions",
    rotten: true,
    stock: 5,
  };
  let updatedApple4: UpdateFood = {
    id: updatedApple2.id,
    name: null,
    rotten: null,
    stock: null,
  };

  expect(await query.all()).toEqual({ collection: [apple, burger] });
  await query.update("non-existing-id", updatedApple1);
  expect(await query.all()).toEqual({ collection: [apple, burger] });
  await query.update(apple.id, updatedApple1);
  expect(await query.all()).toEqual({ collection: [burger, updatedApple1] });
  await query.update(apple.id, updatedApple2);
  expect(await query.all()).toEqual({ collection: [burger, updatedApple2] });
  await query.update("updated_id", updatedApple3);
  expect(await query.all()).toEqual({ collection: [burger, updatedApple3] });
  await query.update("updated_id", updatedApple4);
  expect(await query.all()).toEqual({ collection: [burger, updatedApple4] });
});
