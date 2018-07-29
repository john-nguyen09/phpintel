import { Symbol, TokenSymbol, Consumer } from "../symbol";

export class NamespaceName extends Symbol implements Consumer {
    public name: string = '';

    consume(other: Symbol) {
        if (other instanceof TokenSymbol) {
            this.name += other.text;

            return true;
        }

        return false;
    }
}