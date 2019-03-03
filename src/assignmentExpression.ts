import * as Parser from "tree-sitter";
import { Variable } from "./variable";
import { TreeAnalyser } from "./treeAnalyser";
import { TypeResolver } from "./typeResolver";

export class AssignmentExpression {
    private _isVariable: boolean = false;
    private _variable: Variable = new Variable();

    constructor(node: Parser.SyntaxNode) {
        let hasEqual = false;

        for (const child of node.children) {
            if (child.type === '=') {
                hasEqual = true;
                continue;
            }

            if (!hasEqual) {
                if (child.type === 'variable_name') {
                    this._variable.name = child.text;
                    continue;
                }
            } else {
                const type = TypeResolver.getNodeType(child);

                if (!type.isEmpty) {
                    this._variable.type.push(type);
                }

                break; // TODO: There can be multiple nodes after the equal sign need to handle this
            }
        }
    }

    get isVariable(): boolean {
        return this._isVariable;
    }

    get variable(): Variable {
        return this._variable;
    }
}