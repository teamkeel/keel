import { test, expect, beforeEach } from "vitest";

import { actions, resetDatabase, models } from "@teamkeel/testing";

beforeEach(resetDatabase);

test("query error", async () => {
  const res = await actions.graphql({
    query: `
        query {
            notathing
        }
    `,
  });

  expect(res).toMatchObject({
    errors: [
      {
        message: `Cannot query field "notathing" on type "Query".`,
      },
    ],
  });
});

test("listing with graph traversal", async () => {
  const macmillan = await models.publisher.create({
    name: "Macmillan",
  });
  const randomHouse = await models.publisher.create({
    name: "Random House",
  });
  const terryPratchett = await models.author.create({
    publisherId: macmillan.id,
    name: "Terry Pratchett",
  });
  const douglasAdams = await models.author.create({
    publisherId: randomHouse.id,
    name: "Douglas Adams",
  });
  const colourOfMagic = await models.book.create({
    title: "Colour of Magic",
    authorId: terryPratchett.id,
  });
  await models.book.create({
    title: "Hitchikers Guide To The Galaxy",
    authorId: douglasAdams.id,
  });

  const res = await actions.graphql({
    query: `
        query ListBooks($input: books_input!) {
            books(input: $input) {
                edges {
                    node {
                        id,
                        title
                        author {
                            id
                            publisher {
                                id
                                name
                            }
                        }
                    }
                }
            }
        }
    `,
    variables: {
      input: {
        where: {
          authorPublisherName: {
            equals: "Macmillan",
          },
        },
      },
    },
  });

  expect(res).toEqual({
    data: {
      books: {
        edges: [
          {
            node: {
              id: colourOfMagic.id,
              title: "Colour of Magic",
              author: {
                id: terryPratchett.id,
                publisher: {
                  id: macmillan.id,
                  name: "Macmillan",
                },
              },
            },
          },
        ],
      },
    },
  });
});
