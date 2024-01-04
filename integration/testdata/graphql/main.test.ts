import { test, expect, beforeEach } from "vitest";
import { actions, resetDatabase } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("list action with has-one relationship", async () => {
  const { token } = await actions.authenticate({
    createIfNotExists: true,
    emailPassword: {
      email: "user@keel.xyz",
      password: "1234",
    },
  });

  const authed = actions.withAuthToken(token);

  const user = await authed.createUser({
    firstName: "John",
    lastName: "Lennon",
  });

  await authed.createBlogPost({
    title: "Why I left The Beatles",
    content: "blah blah blah",
  });

  const resp = await graphql(
    `
      query {
        blogPosts {
          edges {
            node {
              id
              user {
                id
                firstName
              }
            }
          }
        }
      }
    `,
    token
  );

  expect(resp.data.blogPosts.edges.length).toBe(1);
  expect(resp.data.blogPosts.edges[0].node.user).toEqual({
    id: user.id,
    firstName: user.firstName,
  });
});

async function graphql(query, token) {
  const res = await fetch(
    process.env.KEEL_TESTING_ACTIONS_API_URL!.replace("/json", "/graphql"),
    {
      method: "POST",
      body: JSON.stringify({
        query,
      }),
      headers: {
        Authorization: `Bearer ${token}`,
      },
    }
  );
  return res.json();
}
