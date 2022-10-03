import { test, expect, actions, Post, Identity } from '@teamkeel/testing'

const newIdentity = async (email : string) => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: email,
    password: '1234'})

  const { object: identity } = await Identity.findOne({ id: identityId })
  
  return identity
} 

test('same identity permission successful', async () => {
  var identity = await newIdentity('user@keel.xyz')

  const { object: post } = await actions
    .withIdentity(identity)  
    .createPostWithIdentity({ })

  expect(
    await actions
    .withIdentity(identity)  
    .deletePostRequiresSameIdentity({ id: post.id })
  ).notToHaveAuthorizationError()
})

test('different identity permission failure', async () => {
  var identity1 = await newIdentity('user1@keel.xyz')
  var identity2 = await newIdentity('user2@keel.xyz')

  const { object: post } = await actions
    .withIdentity(identity1)  
    .createPostWithIdentity({ })

  expect(
    await actions
    .withIdentity(identity2)  
    .deletePostRequiresSameIdentity({ id: post.id })
  ).toHaveAuthorizationError(true)
})

test('unset identity permission failure', async () => {
  var identity = await newIdentity('user@keel.xyz')

  const { object: post } = await actions
    .withIdentity(identity)  
    .createPostWithoutIdentity({ })

  expect(
    await actions
    .withIdentity(identity)  
    .deletePostRequiresSameIdentity({ id: post.id })
  ).toHaveAuthorizationError()
})

// todo:  permission test against null.  Requires this fix:  https://linear.app/keel/issue/DEV-195/permissions-support-null-operand-with-identity-type

// todo:  permission test against another identity field.  Requires this fix: https://linear.app/keel/issue/DEV-196/permissions-support-identity-type-operand-with-identity-comparison
