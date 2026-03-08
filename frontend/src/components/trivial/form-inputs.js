import React from "react";
import {loadQueryParams} from "../../api/API";


const INPUT_LABEL = 0
const INPUT_TYPE = 1
const INPUT_ID = 2
const INPUT_MIN = 3
const INPUT_MAX = 4
const INPUT_OPTIONS = 3
const INPUT_ADDFIRSTBLANK = 4

function setFilter(app, filters) {
        let filtersRes = {}
        filters.map((el) => {
            switch (el[INPUT_TYPE]){
                case 'minmax':
                    filtersRes[el[INPUT_ID]+'min'] = document.getElementById(el[INPUT_ID]+'min').value
                    filtersRes[el[INPUT_ID]+'max'] = document.getElementById(el[INPUT_ID]+'max').value
                    break
                case 'fromto':
                    filtersRes[el[INPUT_ID]+'from'] = document.getElementById(el[INPUT_ID]+'from').value
                    filtersRes[el[INPUT_ID]+'to'] = document.getElementById(el[INPUT_ID]+'to').value
                    break
                case 'choose':
                case 'choose_dict':
                case 'date':
                case 'text':
                    if (document.getElementById(el[INPUT_ID]).value !== '') {
                        filtersRes[el[INPUT_ID]] = document.getElementById(el[INPUT_ID]).value
                    }
                    break
            }
        })
        if(document.getElementById('curPage')) {
            document.getElementById('curPage').value = 1
        }
        const params = app.state
        params.page = 1
        params.filters = filtersRes
        document.location.hash = loadQueryParams(params).replace('?','')
        app.setState({filters: filtersRes, reserveData: app.state.data, data: null, page: 1})
    }

export function FilterCard({app, filters}){
    if (filters === null) return

    function dropFilter() {
        filters.map((el) => {
            const element = document.getElementById(el[INPUT_ID])

            switch (el[INPUT_TYPE]){
                case 'minmax':
                    const min = document.getElementById(el[INPUT_ID]+'min')
                    const max = document.getElementById(el[INPUT_ID]+'max')
                    min.value = min.min
                    max.value = max.max
                    break
                case 'fromto':
                    const from = document.getElementById(el[INPUT_ID]+'from')
                    const to = document.getElementById(el[INPUT_ID]+'to')
                    from.value = null
                    to.value = null
                    from.max = null
                    to.min = null
                    break
                case 'choose':
                case 'choose_dict':
                    element.value = element.options[0].value;
                    break
                case 'date':
                    element.value = null
                    break
                case 'text':
                    element.value = ''
                    break
            }
        })

        const params = app.state
        params.filters = {}
        document.location.hash = loadQueryParams(params).replace('?','')
        app.setState({filters: {}, reserveData: app.state.data, data: null})
    }
    const curFilters = app.state.filters || {}

    return (
        <div className={'filterCard'}>
            <Header label={'Фильтр'}/>
            {
                filters.map((el, i) =>
                    <FilterElement app={app} filters={filters} key={el} curFilters={curFilters} data={el} first={i === 0} />
                )
            }
            <Button  label={'Поиск'} onClick={() => setFilter(app, filters)}/>
            <Button  label={'Сброс'} onClick={dropFilter}/>
        </div>
    )
}


export function Header({label}) {
    return (
        <div>
            <div className={'filter_header'}>
                {label}
            </div>
            <p/>
        </div>
    )
}

function FilterElement({app, filters, curFilters, data, first}){
    switch (data[INPUT_TYPE]){
        case 'minmax':
            return makeMinMax('number', data[INPUT_LABEL], data[INPUT_ID], data[INPUT_MIN], data[INPUT_MAX], curFilters)
        case 'fromto':
            return makeDateInterval('date', data[INPUT_LABEL], data[INPUT_ID], curFilters)
        case 'choose':
            return makeChoose(data[INPUT_LABEL], data[INPUT_ID], data[INPUT_OPTIONS], data[INPUT_ADDFIRSTBLANK], curFilters, first)
        case 'choose_dict':
            return makeChooseDict(data[INPUT_LABEL], data[INPUT_ID], data[INPUT_OPTIONS], data[INPUT_ADDFIRSTBLANK], curFilters, first)
        case 'text':
            return makeTextInput(app, filters,'text', data[INPUT_LABEL], data[INPUT_ID], curFilters)
        case 'date':
            return makeTextInput(app, filters,'date', data[INPUT_LABEL], data[INPUT_ID], curFilters)
    }
}

export function makeDateInterval(type, label, id, curFilters) {
    const setFromBorder = () => {
        document.getElementById(id + 'from').max = document.getElementById(id + 'to').value
    }
    const setToBorder = () => {
        document.getElementById(id + 'to').min = document.getElementById(id + 'from').value
    }
    const minDef = curFilters[id + 'from']
    const maxDef = curFilters[id + 'to']
    return (
        <div>
            <div className={'filter_input_tag'}>
                {label}
            </div>
            <div className={'rowFlex'}>
                <div className={'filter_minmax_text first'}>c</div>
                <input className={'filter_base fromTo blockToRight'} type={type} id={id + 'from'} max={maxDef} defaultValue={minDef} onChange={setToBorder}/>
            </div>
            <div className={'rowFlex'}>
                <div className={'filter_minmax_text  first'}> до</div>
                <input className={'filter_base fromTo blockToRight'} type={type} id={id + 'to'} min={minDef} defaultValue={maxDef} onChange={setFromBorder}/>
            </div>
        </div>
    )
}

export function makeMinMax(type, label, id, min, max, curFilters) {

    const minDef = curFilters[id + 'min'] || min
    const maxDef = curFilters[id + 'max'] || max
    return (
        <div>
            <div className={'filter_input_tag'}>
                {label}
            </div>
            <div className={'rowFlex'}>
                <div className={'filter_minmax_text first'}>c</div>
                <input className={'filter_base minmax'} type={type} id={id + 'min'} min={min} max={max} defaultValue={minDef}/>
                <div className={'filter_minmax_text'}> по </div>
                <input className={'filter_base minmax'} type={type} id={id + 'max'} min={min} max={max} defaultValue={maxDef}/>
            </div>
        </div>
    )
}

function makeChoose(label, id, chooses, skipFirst, curFilters, first) {
    let def_value = curFilters[id]
    if (first) {
        def_value = curFilters['id']
    }

    return (
        <div>
            <div className={'filter_input_tag'}>
                {label}
            </div>
            <select className={'filter_base chooser'} id = {id}  defaultValue={def_value}>
                {
                    chooses.map((el, i) =>
                        makeOption(el,i, skipFirst)
                    )
                }
            </select>
        </div>
    )
}

function makeOption(el, i, skipFirst) {
    if (skipFirst && i === 0) return
    return (
        <option key={el + i} value={i}>
            {el}
        </option>
    )
}

function makeChooseDict(label, id, chooses, addBlank, curFilters, first) {
    let def_value = curFilters[id]
    if (first) {
        def_value = curFilters['id']
    }
    if(Object.keys(chooses).length === 0) return
    return (
        <div>
            <div className={'filter_input_tag'}>
                {label}
            </div>
            <select className={'filter_base chooser'} id = {id}  defaultValue={def_value}>
                {
                    makeOption('',0, !addBlank)
                }
                {
                    chooses.map((el) =>
                        makeOption(el.name,el.id, false)
                    )
                }
            </select>
        </div>
    )
}

export function makeTextInput(app, filters, type, label, id, curFilters) {

    let cur = curFilters[id] || ""
    if (cur.charAt(0) === '%'){
        cur = decodeURI(cur)
    }
    return (
        <div>
        <div className={'filter_input_tag'}>
                {label}
            </div>
            <div className={'rowFlex'}>
                <input className={'filter_base text'}  type={type} id={id} defaultValue={cur} onKeyUp={(e) => {
                    if (e.key === 'Enter'){
                        setFilter(app, filters)
                    }
                }}/>
            </div>
        </div>
    )
}

export function Button({label, onClick}) {
    return (
        <button className={'filter_base button'} onClick={onClick}>
            {label}
        </button>
    )
}