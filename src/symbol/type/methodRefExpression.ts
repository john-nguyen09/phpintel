import { CollectionSymbol, Symbol, Consumer, TokenSymbol, ScopeMember } from "../symbol";
import { ClassRef } from "../class/classRef";
import { MethodCall } from "../function/methodCall";
import { TokenKind } from "../../util/parser";
import { Class } from "../class/class";
import { Interface } from "../interface/interface";

export class MethodRefExpression extends CollectionSymbol implements Consumer, ScopeMember {
    public isParentIncluded = true;
    public classRef: ClassRef = new ClassRef();
    public methodCall: MethodCall = new MethodCall();

    private hasColonColon: boolean = false;

    consume(other: Symbol): boolean {
        if (other instanceof TokenSymbol && other.type == TokenKind.ColonColon) {
            this.hasColonColon = true;
            this.methodCall.scope = this.classRef.type;

            return true;
        }

        if (!this.hasColonColon) {
            return this.classRef.consume(other);
        } else {
            return this.methodCall.consume(other);
        }
    }

    get realSymbols(): Symbol[] {
        return [
            this.classRef,
            this.methodCall
        ]
    }

    setScopeClass(scopeClass: Class | Interface) {
        this.classRef.setScopeClass(scopeClass);
    }
}