import { test, expect, actions, Post, logger } from '@teamkeel/testing'

test('List Where filters - using equal operator (string) - filters correctly', async () => {
  await Post.create({ title: 'Fred' })
  await Post.create({ title: 'NotFred' })

  const { collection: response } = await actions.listPostsEqualString({ 
      requiredTitle: "Fred" 
    })

  expect(response.length).toEqual(1)
  expect(response[0].title).toEqual('Fred')
})


test('List Where filters - using not equal operator (string) - filters correctly', async () => {
  await Post.create({ title: 'Fred' })
  await Post.create({ title: 'NotFred' })

  const { collection: response } = await actions.listPostsNotEqualString({ 
      requiredTitle: "Fred"
    })

  expect(response.length).toEqual(1)
  expect(response[0].title).toEqual('NotFred')
})
