import React from 'react';


export function not_found_label() {
    return <div className={'baseError'}> Ничего не найдено </div>
}

export function too_many_requests() {
    return <div className={'baseError'}> Слишком много запросов </div>
}

export function loading_label() {
    return <div className={'baseError'}/>
}
