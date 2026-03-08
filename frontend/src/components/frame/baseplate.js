import React from 'react';

import {Head} from './head';
import {Body} from "./body";
import {Bot} from "./bot";
import {Message} from "./message";


export function Baseplate(app) {

    return (
        <div>
            <div className={'rowFlex'}>
                <div className={'main'}>
                    <Head/>
                    <Body app={app}/>
                    <Bot/>
                    <Message/>
                </div>
            </div>
            <div className={'backImage'}/>
        </div>
    );
}
