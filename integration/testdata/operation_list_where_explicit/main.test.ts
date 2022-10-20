import { test, expect, actions, Post, logger } from '@teamkeel/testing'

test('startsWith finds correct records', async () => {
  await Post.create({ title: 'Fred' })
  await Post.create({ title: 'NotFred' })

  const { collection: response } = await actions.listPostsEqualityString({ 
      values: { requiredTitle: "Fred" }
    })

  expect(response.length).toEqual(1)
  expect(response[0].title).toEqual('Fred')
})
