import { actions, resetDatabase, models } from "@teamkeel/testing";
import { test, expect, beforeEach } from "vitest";


beforeEach(resetDatabase);


test("create - @set with backlinks", async () => {
    const org = await models.organisation.create({ name: "Keel", isActive: true  });
    const identity = await models.identity.create({ email: "keelson@keel.so" });
    const user = await models.user.create({ name: "Keelson", identityId: identity.id, organisationId: org.id });
    const record = await actions.withIdentity(identity).createRecord({ name: "Tax Records"});

    expect(record).not.toBeNull();
    expect(record.ownerId).toEqual(user.id);
    expect(record.organisationId).toEqual(org.id);
    expect(record.isActive).toEqual(true);
});

test("create - @set with backlinks and 1:M nested create", async () => {
    const org = await models.organisation.create({ name: "Keel", isActive: true  });
    const identity = await models.identity.create({ email: "keelson@keel.so" });
    const user = await models.user.create({ name: "Keelson", identityId: identity.id, organisationId: org.id });
    const record = await actions.withIdentity(identity).createRecordWithChildren({ 
        name: "Tax Records",
        children: [
            { name: "VAT" },
            { name: "Income Tax" },
            { name: "PAYE Tax" }
        ]
    });

    expect(record).not.toBeNull();
    expect(record.ownerId).toEqual(user.id);
    expect(record.organisationId).toEqual(org.id);
    expect(record.isActive).toEqual(true);
    expect(record.parentId).toBeNull();

    const children = await models.record.findMany({ where: { parentId: record.id }});
    expect(children).toHaveLength(3);

    expect(children[0].isActive).toEqual(true);
    expect(children[0].ownerId).toEqual(user.id);
    expect(children[0].organisationId).toEqual(org.id);

    expect(children[1].isActive).toEqual(true);
    expect(children[1].ownerId).toEqual(user.id);
    expect(children[1].organisationId).toEqual(org.id);

    expect(children[2].isActive).toEqual(true);
    expect(children[2].ownerId).toEqual(user.id);
    expect(children[2].organisationId).toEqual(org.id);

});

test("create - @set with backlinks and M:1 nested create", async () => {
    const org = await models.organisation.create({ name: "Keel", isActive: true  });
    const identity = await models.identity.create({ email: "keelson@keel.so" });
    const user = await models.user.create({ name: "Keelson", identityId: identity.id, organisationId: org.id });
    const record = await actions.withIdentity(identity).createRecordWithParent({ 
        name: "Tax Records",
        parent: { name: "Operations" }
    });

    expect(record).not.toBeNull();
    expect(record.ownerId).toEqual(user.id);
    expect(record.organisationId).toEqual(org.id);
    expect(record.isActive).toEqual(true);
    expect(record.parentId).not.toBeNull();

    const parent = await models.record.findMany({ where: { id: { equals: record.parentId }}});

    expect(parent).toHaveLength(1);
    expect(parent[0].ownerId).toEqual(user.id);
    expect(parent[0].organisationId).toEqual(org.id);
    expect(parent[0].isActive).toEqual(true);
    expect(parent[0].parentId).toBeNull();
});

