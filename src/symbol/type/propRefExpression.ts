import { CollectionSymbol, Consumer, Symbol, TokenSymbol } from "../symbol";
import { ClassRef } from "../class/classRef";
import { PropertyRef } from "../variable/propertyRef";
import { nonenumerable } from "../../util/decorator";
import { TokenKind } from "../../util/parser";

export class PropRefExpression extends CollectionSymbol implements Consumer {
    public readonly isParentIncluded = true;
    public classRef: ClassRef = new ClassRef();
    public propRef: PropertyRef = new PropertyRef();

    @nonenumerable
    private hasColonColon = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.ColonColon) {
            this.hasColonColon = true;
            this.propRef.scope = this.classRef.type;

            return true;
        }

        if (!this.hasColonColon) {
            return this.classRef.consume(other);
        } else {
            return this.propRef.consume(other);
        }
    }

    get realSymbols(): Symbol[] {
        return [
            this.classRef,
            this.propRef
        ]
    }
}