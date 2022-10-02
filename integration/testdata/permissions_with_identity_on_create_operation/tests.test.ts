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
  const { errors } = await actions
    .withIdentity(await newIdentity())  
    .createPostWithIdentityRequiresSameIdentity({ })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('different identity permission successful', async () => {
  const { errors } = await actions
    .withIdentity(await newIdentity())  
    .createPostWithIdentityRequiresDifferentIdentity({ })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('unset identity permission failure', async () => {
  const { errors } = await actions
    .withIdentity(await newIdentity())  
    .createPostWithoutIdentityRequiresSameIdentity({ })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

// todo:  permission test against null.  Requires this fix:  https://linear.app/keel/issue/DEV-195/permissions-support-null-operand-with-identity-type

// todo:  permission test against another identity field.  Requires this fix: https://linear.app/keel/issue/DEV-196/permissions-support-identity-type-operand-with-identity-comparison

function hasAuthorizationError(errors?): boolean {
  if (errors == null)
    return false;

  var hasError = false
   errors.forEach(function(error) {
    if(error.message == 'not authorized to access this operation') {
      hasError = true
    }
  });
  
  return hasError;
}