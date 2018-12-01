import { Variable } from "./variable";
import { TreeNode, nodeRange } from "../../util/parseTree";
import { PhpDocument } from "../phpDocument";
import { Symbol, TokenSymbol } from "../symbol";
import { TokenKind } from "../../util/parser";
import { FieldGetter } from "../fieldGetter";
import { Expression } from "../type/expression";
import { Location } from "../meta/location";

export class SimpleVariable extends Variable implements FieldGetter {
    private location: Location;

    constructor(public node: TreeNode, public doc: PhpDocument) {
        super('', undefined);

        this.location = new Location(doc.uri, nodeRange(node, doc.text));
        this.expression = new Expression(null, this.doc);
    }

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.VariableName) {
            this.name = other.text;
        } else {
            let result = this.expression.consume(other);
            
            if (this.expression.type != undefined && !this.expression.type.isEmptyName()) {
                this.type.push(this.expression.type);
            }

            return result;
        }

        return false;
    }

    getFields(): string[] {
        return ['name', 'type'];
    }

    getLocation(): Location {
        return this.location;
    }
}