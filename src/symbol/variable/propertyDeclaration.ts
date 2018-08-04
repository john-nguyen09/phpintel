import { CollectionSymbol, Symbol, Consumer, DocBlockConsumer } from "../symbol";
import { Property } from "./property";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";
import { DocBlock } from "../docBlock";
import { VarDocNode } from "../../docParser";
import { Name } from "../../type/name";

export class PropertyDeclaration extends CollectionSymbol implements Consumer, DocBlockConsumer {
    public realSymbols: Property[] = [];
    public modifier: SymbolModifier = null;

    consume(other: Symbol): boolean {
        if (other instanceof MemberModifierList) {
            this.modifier = other.modifier;

            return true;
        } else if (other instanceof Property) {
            other.modifier = this.modifier;

            this.realSymbols.push(other);

            return true;
        }

        return false;
    }

    consumeDocBlock(docBlock: DocBlock) {
        let docAst = docBlock.docAst;

        if (docAst.kind == 'doc') {
            let varDocNodes = docBlock.getNodes<VarDocNode>('var');
            let endIndex = Math.min(this.realSymbols.length, varDocNodes.length);

            for (let symbol of this.realSymbols) {
                symbol.description = docAst.summary;
            }

            for (let i = 0; i < endIndex; i++) {
                if (varDocNodes[i].variable == null) {
                    this.realSymbols[i].type = new Name(varDocNodes[i].type.name);
                    console.log(varDocNodes[i].type);

                    if (varDocNodes[i].description) {
                        this.realSymbols[i].description = varDocNodes[i].description;
                    }
                } else {
                    let docVarName = '$' + varDocNodes[i].variable;

                    for (let variable of this.realSymbols) {
                        if (variable.name == docVarName) {
                            variable.type = new Name(varDocNodes[i].type.name);

                            if (varDocNodes[i].description) {
                                variable.description = varDocNodes[i].description;
                            }
                            break;
                        }
                    }
                }
            }
        }
    }
}
