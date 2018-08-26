import { CollectionSymbol, Symbol, Consumer, DocBlockConsumer } from "../symbol";
import { Property } from "./property";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";
import { DocBlock } from "../docBlock";
import { VarDocNode, DocNodeKind, toTypeName } from "../../util/docParser";
import { TypeName } from "../../type/name";

export class PropertyDeclaration extends CollectionSymbol implements Consumer, DocBlockConsumer {
    public realSymbols: Property[] = [];
    public modifier: SymbolModifier;

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
        let properties = this.realSymbols;

        if (docAst.kind == 'doc') {
            let varDocNodes = docBlock.getNodes<VarDocNode>(DocNodeKind.Var);
            let endIndex = Math.min(properties.length, varDocNodes.length);

            for (let symbol of properties) {
                symbol.description = docAst.summary;
            }

            for (let i = 0; i < endIndex; i++) {
                let target: Property | null = null;

                if (varDocNodes[i].variable == null) {
                    target = properties[i];
                } else {
                    let docVarName = '$' + varDocNodes[i].variable;

                    for (let property of properties) {
                        if (property.name == docVarName) {
                            target = property;
                        }
                    }
                }

                if (target != null) {
                    let typeName = toTypeName(varDocNodes[i].type);

                    if (typeName != null) {
                        target.type.push(typeName);
                    }

                    if (varDocNodes[i].description) {
                        target.description = varDocNodes[i].description;
                    }
                }
            }
        }
    }
}
