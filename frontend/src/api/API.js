import {setErrorMessage, setSuccessMessage} from "../components/frame/message";

export const backendURL = 'http://'+document.location.hostname+':8000/api/v1';

function workWithPostPutResult(app, response, success, donext) {
    if (response.ok) {
        setSuccessMessage(app, success)
        response.json().then((response) => {
            donext(response)
        })
    } else {
        response.json().then(({error}) => {
            setErrorMessage(app, errorTranslation(error))
        })
    }
}

export function PUTRequest(app, address, body, success, then){
    POSTPUTRequest(app, 'PUT', address, body, success, then)
}

export function POSTRequest(app, address, body, success, then){
    POSTPUTRequest(app, 'POST',  address, body, success, then)
}

export function DELETERequest(app, address, body, success, then){
    POSTPUTRequest(app, 'DELETE',  address, body, success, then)
}

export function POSTPUTRequest(app,method, address, body, success, donext) {
    fetch(
        backendURL + address, {
            method: method,
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
                'Authorization': localStorage.getItem('key_to_my_heart')
            },
            body: JSON.stringify(body)
        }
    ).then(response => {
        if(response.status === 401){
            return
        }
        workWithPostPutResult(app, response, success, donext)
    })
}

function add_param(params, param){
    params += (params.length === 0 ? '?' : '&') + param
    return params
}

export function loadQueryParams({page, sortIndex, sortOrder, filters, exclude}) {
    let params = ""
    if (page > 1)
        params = add_param(params, 'page=' + page)
    if (sortIndex !== null && sortIndex !== undefined) {
        const prefix = (sortOrder !== null && sortOrder === 'DESC') ? '-' : ''
        params = add_param(params, 'sortIndex=' + prefix + sortIndex)
    }
    if(filters !== null && filters !== undefined) {
        for (const [key, value] of Object.entries(filters)) {
            if (exclude === undefined || !exclude.includes(key)) {
                if(key !== '') {
                    params = add_param(params, key + '=' + value)
                }
            }
        }
    }
    return params
}

export function GETRequest(app, address, params, set) {
    const paramsRow = loadQueryParams(params)
    paramsRow.replace('#','?')
    fetch(
        backendURL + address + paramsRow, {
            headers: {
                'Content-Type': 'application/json;charset=utf-8',
            },
        }
    ).then(response => {
        if(response.status === 401){
            return null
        }
        if(response.status === 404){
            return 'not found'
        } else if(response.status === 500){
            return 'not found'
        }
        return response.json()
    }).then((data) => {
        if(data === null){
            return
        }
        set(data)
    })
}

function errorTranslation(error){
    switch (error){
        case 'User name too big':
            return 'Имя пользователя должно быть не длиннее 20 символов'
        case 'Wrong user name or password':
            return 'Неверное имя пользователя или пароль'
        case 'No restore code':
            return 'Не введён код восстановления пароля'
        case 'Admin have no right to change password':
            return 'Админы пароли восстанавливают по-другому'
        case 'Wrong user name':
        case 'User with this name is not exists':
            return 'Пользователь с таким именем не существует'
        case 'User with this name already exists':
            return 'Пользователь с таким именем уже существует'
        case 'Bad email':
            return "Некорректный адрес электронной почты"
        case 'Password should contain at least 1 number':
            return "Пароль не содержит цифру"
        case 'Password too short':
            return "Пароль слишком короткий"
        case 'Password too short, Password should contain at least 1 number':
            return "Пароль слишком короткий и не содержит цифру"
        case 'No password':
            return 'Пароль не введён'
        case 'User not found':
            return 'Неверное имя пользователя или пароль'
        case 'No user name':
            return 'Имя пользователя не введено'
        default:
            return 'Непредвиденная ошибка'
    }
}