import { GetBookAfterQuery } from "@teamkeel/sdk";

const naughtyWords = ["crypto", "blockchain"];

// This function is testing that the afterQuery hook in a get function can be used to modify the returned records
export default GetBookAfterQuery({
  afterQuery(ctx, inputs, record) {
    return {
      ...record,
      title: record.title
        .split(" ")
        .map((x) => {
          if (naughtyWords.includes(x)) {
            return x[0] + "*".repeat(x.length - 2) + x[x.length - 1];
          }
          return x;
        })
        .join(" "),
    };
  },
});
