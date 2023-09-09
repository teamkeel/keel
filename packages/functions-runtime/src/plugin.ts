
import { RawNode,Kysely,sql,SelectExpression,QueryNode,OperationNodeTransformer, ReferenceNode,InsertQueryBuilder, InsertQueryNode, ReturningNode, SelectionNode, UnknownRow, QueryResult, PluginTransformResultArgs, KyselyPlugin, PluginTransformQueryArgs, ColumnNode, ValueNode, RootOperationNode, ColumnUpdateNode } from "kysely"



// class Transformer extends OperationNodeTransformer {
//   protected override transformReturning(node: ReturningNode): ReturningNode {
//     node = super.transformReturning(node);

//     return {
//       ...node,
//       selections: 
//      // name: this.#snakeCase(node.name),
//     }
//   }


// }

export class SelectMyFuncPlugin implements KyselyPlugin {

  transformQuery(args: PluginTransformQueryArgs): RootOperationNode {
      if (args.node.kind === "InsertQueryNode") {
        const returning: ReturningNode | undefined = args.node.returning;
        const selection: SelectionNode[] = [];

        if (returning) {
            selection.push(...returning.selections);
        }

          //   const rawNode = sql.raw(`set_identity_id('12322')`).toOperationNode();
            // const funcSelect = SelectionNode.create(rawNode)
        console.log(selection);

        return {
          ...args.node,
          returning: returning,
        };
        
        // return {
        //   ...QueryNode.cloneWithReturning(
        //     args.node, 
        //     selection
        //   )
        // }
      }
  
      return args.node;
    }
  
    transformResult(
      args: PluginTransformResultArgs
    ): Promise<QueryResult<UnknownRow>> {
      return Promise.resolve(args.result);
    }
  }