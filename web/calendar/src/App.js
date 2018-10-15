import React, {Component} from 'react';

import Calendar from './Calendar.js';
import logo from './newshound_logo.png';

import './App.css';

class App extends Component {
  render() {
    return (
      <div className="App">
        <header className="App-header">
            <img src={logo} className="App-logo" alt="logo" />
        </header>
        <Calendar />
      </div>
    );
  }
}

export default App;
