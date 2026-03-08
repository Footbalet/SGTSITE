import React from "react";
import {make_cell} from "./cell";
import {loading_label, not_found_label} from "./error";

const WORD_HEADER = 0
const WORD_NAME = 1
const WORD_CELLTYPE = 2
const ACTION_HEADER = 0
const ACTION_METHOD = 1

export function Card({app, data, header_text, words, actions}) {
    if (data === undefined) {
        return loading_label()
    }

    if (data === 'not found') {
        return not_found_label()
    }
    return (
        <div className={'style_card_base'}>
            <div>
                <div className={'style_card_head'}>
                    {data[header_text] ? data[header_text] : header_text}
                </div>
                {
                    words.map((el, i) =>
                        <div key={el} className={i % 2 === 0 ? 'rowFlex odd' : 'rowFlex'}>
                            <div className={'style_card_row_part'}>
                                {el[WORD_HEADER]}
                            </div>
                            <div className={'style_card_row_part right'}>
                                {make_cell(app, data, el[WORD_CELLTYPE], el[WORD_NAME])}
                            </div>
                        </div>
                    )
                }
                {
                    actions.map((el) =>
                        <button className={'style_card_button'} key={el} onClick={el[ACTION_METHOD]}>
                            {el[ACTION_HEADER]}
                        </button>
                    )
                }
            </div>
        </div>
    )
}