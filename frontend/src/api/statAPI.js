import {GETRequest} from "./API";


export function getGeneralStat(app, page, sortIndex, sortOrder, filters) {
    GETRequest(
        app,
        '/games/stats',
        {page, sortIndex, sortOrder, filters},
        (data) => app.setState({oneData: data}),
    )
}

export function getActiveGames(app, page) {
    GETRequest(
        app,
        '/games/active',
        {page},
        (data) => app.setState({data: data}),
    )
}

