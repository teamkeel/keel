import { test, expect, Actions, Post } from '@teamkeel/testing'

test('create action', async () => {
  // todo: /Users/adambull/dev/keel/runtime/actions/create.go:46 ERROR: null value in column "title" of relation "post" violates not-null constraint (SQLSTATE 23502)
  const { object: createdPost } = await Actions.createPost({ title: 'foo', subTitle: 'abc' })
  expect.equal(createdPost.title, 'foo')
})

test('create action (unrecognised fields)', async () => {
  const { object: createdPost } = await Actions.createPost({ unknown: 'foo' })

  // todo: replace with errors once we populate them
  expect.equal(createdPost, null)
})

test('get action', async () => {
  const { object: post } = await Actions.createPost({ title: 'foo', subTitle: 'bcd' })
  const { object: fetchedPost } = await Actions.getPost({ id: post.id })

  expect.equal(fetchedPost.id, post.id)
})

// This test verifies that you can't fetch by a field not specified in action inputs
test('get action (non unique)', async () => {
  const { object: post } = await Actions.createPost({ title: 'foo', subTitle: 'cbd' })
  const { object: fetchedPost } = await Actions.getPost({ title: post.title })

  // todo: until we return errors for 404s, we want to assert
  // that nothing is returned
  expect.equal(fetchedPost, null)
})

test('list action - equals', async () => {
  await Post.create({ title: 'apple', subTitle: 'def' })
  await Post.create({ title: 'apple', subTitle: 'efg' })

  const { collection } = await Actions.listPosts({ title: { equals: 'apple' } })

  expect.equal(collection.length, 2)
})

test('list action - contains', async () => {
  await Post.create({ title: 'banan', subTitle: 'fgh' })
  await Post.create({ title: 'banana', subTitle: 'ghi' })

  const { collection } = await Actions.listPosts({ title: { contains: 'ana' } })

  expect.equal(collection.length, 2)
})

test('list action - startsWith', async () => {
  await Post.create({ title: 'adam', subTitle: 'hij' })
  await Post.create({ title: 'adamant', subTitle: 'ijk' })

  const { collection } = await Actions.listPosts({ title: { startsWith: 'adam' } })

  expect.equal(collection.length, 2)
})

test('list action - endsWith', async () => {
  await Post.create({ title: 'star wars', subTitle: 'jkl' })
  await Post.create({ title: 'a post about star wars', subTitle: 'klm' })

  const { collection } = await Actions.listPosts({ title: { endsWith: 'star wars' } })

  expect.equal(collection.length, 2)
})

test('list action - oneOf', async () => {
  await Post.create({ title: 'pear', subTitle: 'lmn' })
  await Post.create({ title: 'mango', subTitle: 'mno' })

  const { collection } = await Actions.listPosts({ title: { oneOf: ['pear', 'mango'] } })

  expect.equal(collection.length, 2)
})

test('delete action', async () => {
  const { object: post } = await Post.create({ title: 'pear', subTitle: 'nop' })

  const { success } = await Actions.deletePost({ id: post.id })

  expect.equal(success, true)
})

test('delete action (other unique field)', async () => {
  const { object: post } = await Post.create({ title: 'pear', subTitle: 'nop' })

  const { success } = await Actions.deletePostBySubTitle({ subTitle: post.subTitle })

  expect.equal(success, true)
})
