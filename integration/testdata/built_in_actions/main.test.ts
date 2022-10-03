import { test, expect, actions, Post } from '@teamkeel/testing'

test('create action', async () => {
  // todo: /Users/adambull/dev/keel/runtime/actions/create.go:46 ERROR: null value in column "title" of relation "post" violates not-null constraint (SQLSTATE 23502)
  const { object: createdPost } = await actions
    .createPost({ title: 'foo', subTitle: 'abc' })

  expect(createdPost.title).toEqual('foo')
})

test('create action (unrecognised fields)', async () => {
  const { object: createdPost } = await actions
    .createPost({ unknown: 'foo' })

  expect(createdPost).toBeEmpty()
})

test('get action', async () => {
  const { object: post } = await actions
    .createPost({ title: 'foo', subTitle: 'bcd' })

  const { object: fetchedPost } = await actions
    .getPost({ id: post.id })

  expect(fetchedPost.id).toEqual(post.id)
})

// This test verifies that you can't fetch by a field not specified in action inputs
test('get action (non unique)', async () => {
  const { object: post } = await actions
    .createPost({ title: 'foo', subTitle: 'cbd' })

  const { object: fetchedPost } = await actions
    .getPost({ title: post.title })

  // todo: until we return errors for 404s, we want to assert
  // that nothing is returned
  expect(fetchedPost).toBeEmpty()
})

test('list action - equals', async () => {
  await Post.create({ title: 'apple', subTitle: 'def' })
  await Post.create({ title: 'apple', subTitle: 'efg' })

  const { collection } = await actions
    .listPosts({ title: { equals: 'apple' } })

  expect(collection.length).toEqual(2)
})

test('list action - contains', async () => {
  await Post.create({ title: 'banan', subTitle: 'fgh' })
  await Post.create({ title: 'banana', subTitle: 'ghi' })

  const { collection } = await actions
    .listPosts({ title: { contains: 'ana' } })

  expect(collection.length).toEqual(2)
})

test('list action - startsWith', async () => {
  await Post.create({ title: 'adam', subTitle: 'hij' })
  await Post.create({ title: 'adamant', subTitle: 'ijk' })

  const { collection } = await actions
    .listPosts({ title: { startsWith: 'adam' } })

  expect(collection.length).toEqual(2)
})

test('list action - endsWith', async () => {
  await Post.create({ title: 'star wars', subTitle: 'jkl' })
  await Post.create({ title: 'a post about star wars', subTitle: 'klm' })

  const { collection } = await actions
    .listPosts({ title: { endsWith: 'star wars' } })

  expect(collection.length).toEqual(2)
})

test('list action - oneOf', async () => {
  await Post.create({ title: 'pear', subTitle: 'lmn' })
  await Post.create({ title: 'mango', subTitle: 'mno' })

  const { collection } = await actions
    .listPosts({ title: { oneOf: ['pear', 'mango'] } })

  expect(collection.length).toEqual(2)
})

test('delete action', async () => {
  const { object: post } = await Post.create({ title: 'pear', subTitle: 'nop' })

  const { success } = await actions
    .deletePost({ id: post.id })

  expect(success).toEqual(true)
})

test('delete action (other unique field)', async () => {
  const { object: post } = await Post.create({ title: 'pear', subTitle: 'nop' })

  const { success } = await actions
    .deletePostBySubTitle({ subTitle: post.subTitle })

  expect(success).toEqual(true)
})

test('update action', async () => {
  const { object: post } = await Post.create({ title: 'watermelon', subTitle: 'opm' })

  const { object: updatedPost } = await actions
    .updatePost({ where: { id: post.id }, values: { title: 'big watermelon' }})

  expect(updatedPost.title).toEqual('big watermelon')
})