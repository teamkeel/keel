import { test, expect, actions, Post, Author, Publisher } from '@teamkeel/testing'

test('get operation where expressions with M:1 relations - all models active - model returned', async () => {
    const { object: publisher } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: author } = await Author.create({ name: 'Keelson', thePublisherId: publisher.id, isActive: true })
    const { object: firstpost } = await Post.create({ title: 'My First Post', theAuthorId: author.id, isActive: true })

    const { object: post } = await actions.getActivePost({ id: firstpost.id })

    expect(post.id).toEqual(firstpost.id)
})

test('get operation where expressions with M:1 relations - post model not active - no records found', async () => {
    const { object: publisher } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: author } = await Author.create({ name: 'Keelson', thePublisherId: publisher.id, isActive: true })
    const { object: firstpost } = await Post.create({ title: 'My First Post', theAuthorId: author.id, isActive: false })

    expect(
        await actions.getActivePost({ id: firstpost.id })
    ).toHaveError({
        message: "no records found for Get() operation",
    })
})

test('get operation where expressions with M:1 relations - nested author model not active - no records found', async () => {
    const { object: publisher } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: author } = await Author.create({ name: 'Keelson', thePublisherId: publisher.id, isActive: false })
    const { object: firstpost } = await Post.create({ title: 'My First Post', theAuthorId: author.id, isActive: true })

    expect(
        await actions.getActivePost({ id: firstpost.id })
    ).toHaveError({
        message: "no records found for Get() operation",
    })
})

test('get operation where expressions with M:1 relations - nested nested publisher model not active - no records found', async () => {
    const { object: publisher } = await Publisher.create({ orgName: 'Keel Org', isActive: false })
    const { object: author } = await Author.create({ name: 'Keelson', thePublisherId: publisher.id, isActive: true })
    const { object: firstpost } = await Post.create({ title: 'My First Post', theAuthorId: author.id, isActive: true })

    expect(
        await actions.getActivePost({ id: firstpost.id })
    ).toHaveError({
        message: "no records found for Get() operation",
    })
})

test('get operation where expressions with 1:M relations - all models active - publisher returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherKeel.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { object: publisher } = await actions.getActivePublisherWithActivePosts({ id: publisherKeel.id })

    expect(publisher.id).toEqual(publisherKeel.id)
})

test('get operation where expressions with 1:M relations - publisher not active - no publisher found', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: false })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherKeel.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    expect(
        await actions.getActivePublisherWithActivePosts({ id: publisherKeel.id })
    ).toHaveError({
        message: "no records found for Get() operation",
    })
})

test('get operation where expressions with 1:M relations - one author active - publisher returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherKeel.id, isActive: false })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { object: publisher } = await actions.getActivePublisherWithActivePosts({ id: publisherKeel.id })

    expect(publisher.id).toEqual(publisherKeel.id)
})

test('get operation where expressions with 1:M relations - active author with inactive posts and inactive autor with active posts - no publisher found', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherKeel.id, isActive: false })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: false })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: false })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    expect(
        await actions.getActivePublisherWithActivePosts({ id: publisherKeel.id })
    ).toHaveError({
        message: "no records found for Get() operation",
    })
})

test('get operation where expressions with 1:M relations - no active posts  - publisher returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherKeel.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: false })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: false })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: false })

    expect(
        await actions.getActivePublisherWithActivePosts({ id: publisherKeel.id })
    ).toHaveError({
        message: "no records found for Get() operation",
    })
})

test('list operation where expressions with M:1 relations - all models active - all models returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: posts } = await actions.listActivePosts({ })

    expect(posts.length).toEqual(3)
})

test('list operation where expressions with M:1 relations - Keel org not active - Weave models returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: false })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: posts } = await actions.listActivePosts({ })

    expect(posts.length).toEqual(1)
})

test('list operation where expressions with M:1 relations - Keelson author not active - Weaveton models returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: false })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: posts } = await actions.listActivePosts({ })

    expect(posts.length).toEqual(1)
})

test('list operation where expressions with M:1 relations - one Keelson post not active - Weaveton models returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: false })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: posts } = await actions.listActivePosts({ })

    expect(posts.length).toEqual(2)
})

test('list operation where expressions with 1:M relations - all models active - all models returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org 2', isActive: true })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org 2', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: publishers } = await actions.listActivePublishersWithActivePosts({ })

    expect(publishers.length).toEqual(2)
})

test('list operation where expressions with 1:M relations - Keel org not active - only Keel returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: false })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: publishers } = await actions.listActivePublishersWithActivePosts({ })

    expect(publishers.length).toEqual(1)
})

test('list operation where expressions with 1:M relations - Keel author not active - Weave org returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: false })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: true })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: publishers } = await actions.listActivePublishersWithActivePosts({ })

    expect(publishers.length).toEqual(1)
})

test('list operation where expressions with 1:M relations - one Keel post not active - all models returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: true })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: false })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: publishers } = await actions.listActivePublishersWithActivePosts({ })

    expect(publishers.length).toEqual(2)
})

test('list operation where expressions with 1:M relations - all Keel posts not active - Weave org returned', async () => {
    const { object: publisherKeel } = await Publisher.create({ orgName: 'Keel Org', isActive: true })
    const { object: publisherWeave } = await Publisher.create({ orgName: 'Weave Org', isActive: true })
    const { object: author1 } = await Author.create({ name: 'Keelson', thePublisherId: publisherKeel.id, isActive: true })
    const { object: author2 } = await Author.create({ name: 'Weaveton', thePublisherId: publisherWeave.id, isActive: true })
    const { object: post1 } = await Post.create({ title: 'Keelson First Post', theAuthorId: author1.id, isActive: false })
    const { object: post2 } = await Post.create({ title: 'Keelson Second Post', theAuthorId: author1.id, isActive: false })
    const { object: post3 } = await Post.create({ title: 'Weaveton First Post', theAuthorId: author2.id, isActive: true })

    const { collection: publishers } = await actions.listActivePublishersWithActivePosts({ })

    expect(publishers.length).toEqual(1)
})
