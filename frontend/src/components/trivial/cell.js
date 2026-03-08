import React from "react";


function make_content_cell(data){
    let content = data['content']
    if (content.length > 100) {
        content = content.substring(0, 100) + '...';
    }
    return content
}

export function make_cell(app, data, cell_type, cell_name) {
    switch (cell_type) {
        case 'content':
            return make_content_cell(data)
        case 'date':
            return make_date_cell(data, cell_name)
        case 'new_link':
            return make_news_name_link('news_card', data.id, data.title, false, '')
        case 'textinput':
            return make_text_input(data, cell_name)
        case 'dateinput':
            return make_date_input(data, cell_name)
        case 'multilineinput':
            return make_textarea_input(data, cell_name)
        case 'numerinput':
            return make_numer_input(data, cell_name)
        case 'theme':
            return make_theme(data)
        case 'release_download':
            return <button className={'cell_btn'} onClick={() => download_game(data.id)}>{cell_name}</button>
        case 'release_link':
            return <LinkCell href={'release_card/?id='+data.id} label={data.version} />
        default:
            return make_common_cell(data, cell_name)
    }
}

async function download_game(id) {
    const response = await fetch(
        `http://192.144.56.245:8000/api/v1/releases/${id}/download`,
        {
            method: 'GET',
        }
    );

    if (!response.ok) {
        throw new Error('Download failed');
    }
    const blob = await response.blob();
    const contentDisposition = response.headers.get('content-disposition');
    let filename = `LiveFire League.exe`;

    if (contentDisposition) {
        const filenameMatch = contentDisposition.match(/filename="(.+)"/);
        if (filenameMatch && filenameMatch[1]) {
            filename = filenameMatch[1];
        }
    }

    // Создаем ссылку для скачивания
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = filename;
    document.body.appendChild(link);
    link.click();

    // Очистка
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
}

function make_theme(data){
    switch (data.theme){
        case '1': return "Разработка"
        case '2': return "Бета-тест"
        case '3': return "Обновления"
        case '4': return "Прочее"
    }
}

function make_news_name_link(href, id, name, usePlaceholder, placeholder){
    if (usePlaceholder){
        return placeholder
    }
    return <LinkCell href={href+'/?id='+id} label={name} />
}

export function LinkCell({href, label}) {
    return <a href={ '/'+href}>{label}</a>
}

function make_date_cell(data, cell_name) {
    if (data[cell_name] === null)
        return ''
    const date = data[cell_name].split('T')[0].split('-')
    return date[2]+'-'+date[1]+'-'+date[0]
}

function make_common_cell(data, cell_name) {
    return data[cell_name] === null || data[cell_name] === undefined  || data[cell_name] === '' ? ' ' : data[cell_name];
}

function make_text_input(data, cell_name) {
    return <input type={'text'}  className={'textInput'} id = {cell_name} defaultValue={data[cell_name]} />
}

function make_date_input(data, cell_name) {
    let date = data[cell_name] !== null ? data[cell_name].split('T')[0] : ''
    return <input className={'dateInput'} type={'date'} id={cell_name} defaultValue={date}/>
}

function make_numer_input(data, cell_name) {
    return <input className={'numberInput'} type={'number'} id={data.id} max={99} min={1} defaultValue={data[cell_name]}/>
}

function make_textarea_input(data, cell_name) {
    return <textarea className={'multiline'} id={cell_name} defaultValue={data[cell_name]}/>
}

