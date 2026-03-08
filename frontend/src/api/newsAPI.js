import {DELETERequest, GETRequest, POSTRequest, PUTRequest} from "./API";


// Получение всех новостей

export function getAllNews(app, page, sortIndex, sortOrder, filters) {
    GETRequest(
        app,
        '/news/search',
        {page, sortIndex, sortOrder, filters},
        (data) => app.setState({data: data}),
    )
}

export function getDetailNews(app, index) {
    GETRequest(
        app,
        '/news/' + index,
        {},
        (data) => app.setState({oneData: data}),
    )
}

export function createNews(app, theme, title, content) {
    POSTRequest(
        app,
        '/news/',
        {theme, title, content},
        'Новость добавлена',
        () => {}
    )
}

export function updateNews(app, id, theme, title, content) {
    PUTRequest(
        app,
        '/news/'+id,
        {theme, title, content},
        'Новость изменена',
        () => {}
    )
}

export function deleteNews(app, index) {
    DELETERequest(
        app,
        '/news/'+ index,
        {},
        'Новость удалена',
        () => {}
    )
}
