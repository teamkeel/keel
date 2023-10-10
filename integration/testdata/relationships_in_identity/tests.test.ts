import { actions, models, resetDatabase } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";

beforeEach(resetDatabase);


test("asc", async () => {
    const maryIdentity = await models.identity.create({ email: "mary@weave.so "});
    const johnIdentity = await models.identity.create({ email: "john@weave.so "});
    const arthurIdentity = await models.identity.create({ email: "arthur@weave.so "});

    const mary = await actions.withIdentity(maryIdentity).newAccount({ username: "mary77" });
    const john = await actions.withIdentity(johnIdentity).newAccount({ username: "johndoe" });
    const arthur = await actions.withIdentity(arthurIdentity).newAccount({ username: "art" });

    await actions.follow({ 
      account: { id: mary.id }, 
      follower: { id: john.id } 
    });

    await actions.follow({ 
      account: { id: mary.id }, 
      follower: { id: arthur.id } 
    });

    // const johnIsFollowing = await actions.withIdentity(johnIdentity).accountsFollowing();
    // expect(johnIsFollowing.results).toHaveLength(1);
    // expect(johnIsFollowing.results[0].username).toEqual("mary77");

    const johnIsNotFollowing = await actions.withIdentity(johnIdentity).accountsNotFollowing();
    expect(johnIsNotFollowing.results).toHaveLength(1);
    console.log(johnIsNotFollowing.results);
    expect(johnIsNotFollowing.results[0].username).toEqual("art");

    // const authorIsFollowing = await actions.withIdentity(arthurIdentity).accountsFollowing();
    // expect(authorIsFollowing.results).toHaveLength(1);
    // expect(authorIsFollowing.results[0].username).toEqual("mary77");

    // const authorIsNotFollowing = await actions.withIdentity(arthurIdentity).accountsNotFollowing();
    // expect(authorIsNotFollowing.results).toHaveLength(1);
    // expect(authorIsNotFollowing.results[0].username).toEqual("johndoe");

    // const maryIsFollowing = await actions.withIdentity(maryIdentity).accountsFollowing();
    // expect(maryIsFollowing.results).toHaveLength(0);

    // const maryIsNotFollowing = await actions.withIdentity(maryIdentity).accountsNotFollowing();
    // expect(maryIsNotFollowing.results).toHaveLength(2);
});
