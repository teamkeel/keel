import { test, expect, actions, Post, logger } from '@teamkeel/testing'


test('get action with no permissions should fail', async () => {

  const returned = await actions
    .getPost({ id: "unused id" })

    console.log("XXXX response is: ", returned)
})