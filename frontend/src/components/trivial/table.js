import React from "react";
import {loading_label, not_found_label, too_many_requests} from "./error";
import {make_cell} from "./cell";
import {FilterCard, Header} from "./form-inputs";
import {loadQueryParams} from "../../api/API";

export function Table({app, data, tabs, group_by=null, filters, page=app.state.page || 1, large, actions, drawErrors=true, card=null, forceDrawCard=false, missedRows=[], isColorNeeded=(data)=>{return false}}) {

    large = large || false
    actions = actions || null

    if (data === null && app.state.reserveData !== undefined) {
        data = app.state.reserveData
    }
    const activeTab = app.state.curTab || 0
    if (card !== null) {
        if ((data === undefined || data === null || data.count === 0) && !drawErrors && forceDrawCard) {
            return <div>
                {card}
                {
                    actions === null ? '' : <ActionCard app={app} actions={actions} className={'actionCard'}/>
                }
            </div>
        }
        if (data === null) {
            return drawErrors ? loading_label() : ''
        } else if (data.error){
            return drawErrors ? too_many_requests() : ''
        } else if (data.count === 0) {
            return <div className={'rowFlex'}>
                {drawErrors ? not_found_label() : ''}
                <div>
                    {card}
                    {
                        actions === null ? '' : <ActionCard app={app} actions={actions} className={'actionCard'}/>
                    }
                </div>
            </div>
        }
    }

    return (
        <div>
            {
                (data === null || data === undefined || data.count === 0) ? '' :
                    <Tabs app={app} tabs={tabs} activeTab={activeTab}/>
            }
            <div className={'rowFlex'}>
                <TableBase app={app} tabs={tabs} data={data} group_by={group_by} page={page} large={large}
                           drawErrors={drawErrors} missedRows={missedRows} isColorNeeded={isColorNeeded}/>
                {
                    filters === null ? '' :
                    <div className={'filterCardBase'}>
                        <FilterCard app={app} filters={filters}/>
                        {
                            actions === null ? '' : <ActionCard app={app} actions={actions}/>
                        }
                    </div>
                }
                {
                    card === null ? '' :
                        <div>
                            {card}
                            {
                                actions === null ? '' : <ActionCard app={app} actions={actions}/>
                            }
                        </div>
                }
            </div>
        </div>
    )
}

function TableBase({app, tabs, data, group_by, page, large, drawErrors, missedRows, isColorNeeded=()=>{return false}}) {
    if (data === null) {
        return drawErrors ? loading_label() : ''
    } else if (data.error){
        return drawErrors ? too_many_requests() : ''
    } else if (data === undefined || data.count === 0) {
        return drawErrors ? not_found_label() : ''
    }
    const pages = app.state.nopagination ? 1 : Math.floor((data.count - 1) / 30) + 1
    const activeTab = app.state.curTab || 0
    const tableTab = tabs[activeTab][1]

    const wides = []
    tableTab.map(el => {
        wides.push(el[3])
    })

    return (
        <div className={'style_table_base'}>
            <div className={'style_table_body'}>
                <Heads app={app} wides={wides} tableTab={tableTab} large={large} missedRows={missedRows}/>
                <Rows app={app} wides={wides} data={data} group_by={group_by} tableTab={tableTab} page={page} large={large} missedRows={missedRows} isColorNeeded={isColorNeeded}/>
                <PageRow app={app} data={data} page={page} pages={pages}/>
            </div>
        </div>
    )
}

function Tabs({app, tabs, activeTab}) {
    if(tabs.length === 1) return
    function setTab(i) {
        const body = document.getElementById('table_rows')
        body.style.animation = 'none'
        setTimeout(function() {
            body.style.animation = '';
        }, 10);
        app.setState({curTab: i})
    }

    return (
        <div className={'style_table_tab_row'}>
            {
                tabs.map((el, i) =>
                    <button className={activeTab === i ? 'style_table_tab active' : 'style_table_tab'} key={el}
                            onClick={() => setTab(i)}
                    >
                        {el[0]}
                    </button>
                )
            }
            <div className={'style_table_tab'} style={{opacity: 0}}>

            </div>
        </div>
    )
}

function Heads({app, tableTab, wides, missedRows}) {

    let sort = app.state.sortIndex || tableTab[0][1]
    let order = app.state.sortOrder || 'ASC'


    function setSort(value) {
        const orderOld = app.state.sortOrder || 'ASC'
        let order = 'ASC'
        let newIndex = value
        if (app.state.sortIndex !== null && app.state.sortIndex === value) {
            if (newIndex === tableTab[0][1]) {
                if (orderOld === 'ASC') {
                    order = 'DESC'
                } else {
                    order = 'ASC'
                }
            } else {
                if (orderOld === 'ASC') {
                    order = 'DESC'
                } else {
                    newIndex = null
                    order = null
                }
            }
        }
        const params = app.state
        params.sortIndex = newIndex
        params.sortOrder = order
        document.location.hash = loadQueryParams(params).replace('?', '')
        app.setState({sortIndex: newIndex, sortOrder: order, reserveData: app.state.data, data: null})
    }

    function setToolTip(e, text) {
        if (text !== '' && text !== undefined) {
            const tooltip = document.getElementById('tooltip')
            tooltip.innerText = text
            moveToolTip(e, text)
            tooltip.style.opacity = '100%'
        }
    }

    function moveToolTip(e, text) {
        if (text !== '') {
            const tooltip = document.getElementById('tooltip')
            const x = e.clientX;
            const y = e.clientY;
            tooltip.style.left = x + document.documentElement.scrollLeft + 'px'
            tooltip.style.top = y + 20 + document.documentElement.scrollTop +'px'
        }
    }

    function hideToolTip() {
        const tooltip = document.getElementById('tooltip')
        tooltip.style.opacity = '0%'
    }
    return (
        <div>
            <div className={'rowFlex'} style={{'background-color': '#040404ff'}}>
                {
                    tableTab.map((el, i) => missedRows.includes(el[1]) ? "" :
                        <div key={el} style={{width: wides[i]+'%'}}
                             onMouseOver={(e) => {setToolTip(e, el[4])}}
                             onMouseMove={(e) => {moveToolTip(e, el[4])}}
                             onMouseLeave={hideToolTip}
                             >
                            <button
                                className={el[5] !== false ? 'style_table_head' : 'style_table_head disabled'}
                                onClick={() => {
                                    if (el[5] !== false)
                                        setSort(el[1])
                                }}>
                                <div className={'rowFlex full'}>
                                <div style={{'color': 'white'}}>
                                    {el[0]}
                                </div>
                                <div className={sort === el[1] ? (order === 'ASC' ? 'sortTable ascended' : 'sortTable descended') : ''} />
                                </div>
                            </button>
                        </div>
                    )
                }
            </div>
            <div id='tooltip' className={'moveableDiv'} />
        </div>
    )
}

function Rows({app, data, group_by, tableTab, page, wides, missedRows, isColorNeeded=(data)=> {return false}}) {

    if (group_by === null) {
        return (
            <div id={'table_rows'} className={'table_rows'}>
                {data.results.map((el, i) =>  <Row key={el + i} app={app} wides={wides} el={el} i={i} tableTab={tableTab} page={page} missedRows={missedRows} isColorNeeded={isColorNeeded}></Row>
                )}
            </div>
        )
    } else {
        if (group_by === 'position') {
            const group_gk = []
            const group_fp = []
            data.results.map((el) => {
                if (el.position === 1) {
                    group_gk.push(el)
                } else {
                    group_fp.push(el)
                }
            })
            return (
                <div>
                    <div id={'table_rows'} className={'table_rows'}>
                        {group_gk.map((el, i) => <Row key={el + i} app={app} wides={wides} el={el} i={i} tableTab={tableTab} page={page} missedRows={missedRows}></Row>
                        )}
                    </div>
                    <div className={'tableLine'} />
                    <div id={'table_rows'} className={'table_rows'}>
                        {group_fp.map((el, i) => <Row key={el + i} app={app} wides={wides} el={el} i={i} tableTab={tableTab} page={page} missedRows={missedRows}></Row>
                        )}
                    </div>
                </div>
            )
        }
    }
}

function Row({app, el, tableTab, page, i, wides, missedRows, isColorNeeded=(data)=>{return false}}) {
    let colored = ''
    try {
        colored = isColorNeeded(el) ? ' yellow' : ''
    } catch (e){
        colored = ''
    }
    return (
        <div className={(i % 2 === 0 ? 'rowFlex odd' : 'rowFlex') + colored} style={{borderTop: 'solid 1px #555555'}} key={i}>
            {
                tableTab.map((ele, j) => missedRows.includes(ele[1]) ? "" :
                    <div key={i + ':' + j} style={{width: wides[j]+'%'}}>
                        <div className={'style_table_row_cell'}>
                            {
                                ele[2] === 'index' ?
                                    <div>
                                        {(i + 1 + 30 * (page - 1))}
                                    </div>
                                    :
                                    make_cell(app, el, ele[2], ele[1])
                            }
                        </div>
                    </div>
                )
            }
        </div>
    )
}

function PageRow({app, data, page, pages}) {

    if(pages === 1) return

    function toBegin() {
        if (page > 1)
            setTablePage(1)
    }
    function back() {
        if (page > 1)
            setTablePage(page-1)
    }
    function forward() {
        if (page < pages)
            setTablePage(page+1)
    }
    function toEnd() {
        if (page < pages)
            setTablePage(pages)
    }

    function setPage() {
        const pageInput = document.getElementById('curPage')
        let page = pageInput.value
        if (page > pages){
            page = pages
        } else if (page < 1){
            page = 1
        }
        pageInput.value = page
        setTablePage(parseInt(document.getElementById('curPage').value))
    }

    function setTablePage(newPage){
        const params = app.state
        params.page = newPage
        document.location.hash = loadQueryParams(params).replace('?','')
        app.setState({page: newPage, reserveData: app.state.data, data: null})
        const pageInput = document.getElementById('curPage')
        pageInput.value = newPage
    }

    return (
        <div className={'rowFlex'}>
            <button className={'pages_button begin'} onClick={toBegin}/>
            <button className={'pages_button back'} onClick={back}/>
            <input className={'pages_input'} id={'curPage'} type={"number"} defaultValue={page} min={1} max={pages} onKeyUp={(e) => {
                if (e.code === 'Enter') setPage()}}/>
            <div className={'pages_count'}> {' /' + pages} </div>
            <button className={'pages_button forward'} onClick={forward}/>
            <button className={'pages_button end'} onClick={toEnd}/>
            <div className={'pages_count'}>{' Всего записей: ' + data.count} </div>
        </div>
    )
}

function ActionCard({app, actions, className='filterCard'}) {

    function get_value(filter) {
        return app.state.filters[filter] || ''
    }

    function get_filters(filters) {
        let line = ''
        filters.map((el, i) => {
            if (i > 0) {
                line += '&'
            }
            line += el + '=' + get_value(el)
        })
        return line
    }

    function get_link(mainPath, filters) {
        const filters_part = get_filters(filters)
        return ('/' + mainPath + '#' + filters_part)
    }

    return (
        <div className={className}>
            <Header label={'Переходы'}/>
            {
                actions.map(el => <a key={el} id={el[1]} href={get_link(el[1], el[2])} className={'div_href'}>
                        {el[0]}
                    </a>
                )
            }
        </div>
    )
}