import { Symbol, TokenSymbol, Consumer } from "../symbol";
import { TokenKind } from "../../util/parser";

export class NamespaceName extends Symbol implements Consumer {
    public parts: string[] = [];

    consume(other: Symbol) {
        if (other instanceof TokenSymbol && other.type == TokenKind.Name) {
            this.parts.push(other.text);

            return true;
        }

        return false;
    }

    get name(): string {
        return this.parts.join('\\');
    }

    get fqn(): string {
        return '\\' + this.name;
    }

    get isRoot(): boolean {
        return this.parts.length === 0 || this.fqn === '\\';
    }
}