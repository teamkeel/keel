import { test, expect, actions, Post, Identity } from '@teamkeel/testing'

const newIdentity = async () => {
  const { identityId } = await actions.authenticate({ 
    createIfNotExists: true, 
    email: 'user@keel.xyz',
    password: '1234'})

  const { object: identity } = await Identity.findOne({ id: identityId })
  
  return identity
} 

test('same identity permission successful', async () => {
  expect(
    await actions
    .withIdentity(await newIdentity())  
    .createPostWithIdentityRequiresSameIdentity({ })
  ).notToHaveAuthorizationError()
})

test('different identity permission failure', async () => {
  expect(
    await actions
    .withIdentity(await newIdentity())  
    .createPostWithIdentityRequiresDifferentIdentity({ })
  ).toHaveAuthorizationError()
})

test('unset identity permission failure', async () => {
  expect(
    await actions
    .withIdentity(await newIdentity())  
    .createPostWithoutIdentityRequiresSameIdentity({ })
  ).toHaveAuthorizationError()
})

// todo:  permission test against null.  Requires this fix:  https://linear.app/keel/issue/DEV-195/permissions-support-null-operand-with-identity-type

// todo:  permission test against another identity field.  Requires this fix: https://linear.app/keel/issue/DEV-196/permissions-support-identity-type-operand-with-identity-comparison
