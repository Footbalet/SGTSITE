import {DELETERequest, GETRequest} from "./API";


// Получение всех новостей

export function getAllReleases(app, page, sortIndex, sortOrder, filters) {
    GETRequest(
        app,
        '/releases/',
        {page, sortIndex, sortOrder, filters},
        (data) => app.setState({data: data}),
    )
}

export function getDetailRelease(app, index) {
    GETRequest(
        app,
        '/releases/' + index,
        {},
        (data) => app.setState({oneData: data}),
    )
}

export function deleteRelease(app, index) {
    DELETERequest(
        app,
        '/releases/'+ index,
        {},
        'Релиз удален',
        () => {}
    )
}
