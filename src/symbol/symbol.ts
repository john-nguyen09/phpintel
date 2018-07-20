import { Location } from './meta/location';
import { SymbolKind } from './meta/kind';
import { SymbolModifier } from './meta/modifier';
import { TreeNode } from '../util/parseTree';
import { TokenType, Token } from 'php7parser';

export interface Symbol {
    node: TreeNode;
    consume(other: Symbol): void;
}