import { CollectionSymbol, Symbol, Consumer, DocBlockConsumer } from "../symbol";
import { Property } from "./property";
import { SymbolModifier } from "../meta/modifier";
import { MemberModifierList } from "../class/memberModifierList";
import { DocBlock } from "../docBlock";
import { VarDocNode, DocNodeKind } from "../../docParser";
import { TypeName } from "../../type/name";

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
        let variables = this.realSymbols;

        if (docAst.kind == 'doc') {
            let varDocNodes = docBlock.getNodes<VarDocNode>(DocNodeKind.Var);
            let endIndex = Math.min(variables.length, varDocNodes.length);

            for (let symbol of variables) {
                symbol.description = docAst.summary;
            }

            for (let i = 0; i < endIndex; i++) {
                if (varDocNodes[i].variable == null) {
                    variables[i].type.push(new TypeName(varDocNodes[i].type.name));

                    if (varDocNodes[i].description) {
                        variables[i].description = varDocNodes[i].description;
                    }
                } else {
                    let docVarName = '$' + varDocNodes[i].variable;

                    for (let variable of variables) {
                        if (variable.name == docVarName) {
                            variable.type.push(new TypeName(varDocNodes[i].type.name));

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
