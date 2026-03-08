import React from 'react'


export function setErrorMessage(app, error){
    setMessage('messageBox Error', error)
}

export function setSuccessMessage(app, success){
    setMessage('messageBox', success)
}

function setMessage(className, message){
    const box = document.getElementById('message_box')
    box.style.animation = 'none'
    setTimeout(()=> {
        box.style.animation = ''
    }, 10)
    box.innerHTML = message
    box.className = className
}

export function Message(){
    return(
        <div id={'message_box'} className={'messageBox'} style={{animation:'none'}} />
    )
}