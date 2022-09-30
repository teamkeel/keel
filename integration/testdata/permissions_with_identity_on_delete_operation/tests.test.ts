import { test, expect, actions, Post } from '@teamkeel/testing'

test('same identity permission successful', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithIdentity({ })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostRequiresSameIdentity({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, false)
})

test('different identity permission failure', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithIdentity({ })

  const { errors } = await actions
    .withIdentity('2FQqdDYm47mEjgEGsUTtVYbDmuM')  
    .deletePostRequiresSameIdentity({ id: post.id })

  var authorizationFailed = hasAuthorizationError(errors)
  expect.equal(authorizationFailed, true)
})

test('unset identity permission failure', async () => {
  const { object: post } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .createPostWithoutIdentity({ })

  const { errors } = await actions
    .withIdentity('0ujsszgFvbiEr7CDgE3z8MAUPFt')  
    .deletePostRequiresSameIdentity({ id: post.id })

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