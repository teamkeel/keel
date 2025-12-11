import { ParentTask, tasks } from "@teamkeel/sdk";

export default ParentTask({}, async (ctx, inputs) => {
  // Create child tasks from within the flow using the tasks SDK
  const createdTasks: string[] = [];

  for (let i = 0; i < inputs.childCount; i++) {
    // Use withIdentity to create tasks as the current identity
    const childTask = await tasks.childTask.withIdentity(ctx.identity!).create({
      parentName: inputs.name,
      index: i + 1,
    });
    createdTasks.push(childTask.id);
  }

  return ctx.complete({
    data: {
      parentName: inputs.name,
      childTaskIds: createdTasks,
      childCount: createdTasks.length,
    },
  });
});
