import { test, expect, actions, Post } from '@teamkeel/testing'

test('same identity permission successful', async () => {
  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithIdentityRequiresSameIdentity({ })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('different identity permission successful', async () => {
  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithIdentityRequiresDifferentIdentity({ })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('unset identity permission failure', async () => {
  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithoutIdentityRequiresSameIdentity({ })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

// todo:  add test which test with null.  Requires this fix:  https://linear.app/keel/issue/DEV-195/permissions-support-null-operand-with-identity-type

// todo:  add test which checks against another identity field.  Requires this fix: https://linear.app/keel/issue/DEV-196/permissions-support-identity-type-operand-with-identity-comparison

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