import { CollectionSymbol, Symbol, Consumer, DocBlockConsumer } from "../symbol";
import { Property } from "./property";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";
import { DocBlock } from "../docBlock";

export class PropertyDeclaration extends Symbol implements Consumer, DocBlockConsumer {
    public modifier: SymbolModifier;

    private doc: DocBlock | null = null;

    consume(other: Symbol): boolean {
        if (other instanceof MemberModifierList) {
            this.modifier = other.modifier;

            return true;
        } else if (other instanceof Property) {
            other.modifier = this.modifier;

            if (this.doc !== null) {
                other.consumeDocBlock(this.doc);
            }

            return true;
        }

        return false;
    }

    consumeDocBlock(doc: DocBlock) {
        this.doc = doc;
    }
}
