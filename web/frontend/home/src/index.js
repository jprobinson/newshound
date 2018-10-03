import React from 'react';
import ReactDOM from 'react-dom';

class Feed extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      events: [],
      isLoaded: false,
      error: null,
    };

    const todayDate = new Date();
    const startDate = new Date(
      todayDate.getFullYear(),
      todayDate.getMonth(),
      todayDate.getDate() - 7,
    );
    const today = todayDate.toISOString().split('T')[0];
    const start = startDate.toISOString().split('T')[0];

    this.getEvents(start, today);
  }

  getEvents(start, end) {
    fetch(
      'https://newshound.jprbnsn.com/svc/newshound-api/v1/event_feed/' +
        start +
        '/' +
        end,
    )
      .then(res => res.json())
      .then(
        result => {
          this.setState({
            events: result,
            isLoaded: true,
          });
        },
        error => {
          this.setState({
            isLoaded: true,
            error,
          });
        },
      );
  }

  render() {
    const {error, isLoaded, events} = this.state;
    if (error) {
      return <div>ERROR: {error.message} </div>;
    } else if (!isLoaded) {
      return <div className="feed-container">Loading...</div>;
    } else {
      return (
        <div className="feed-container">
          {events.map(event => (
            <Event {...event}/>
          ))}
        </div>
      );
    }
  }
}

class Event extends React.Component {
  render() {
	console.log(this);
    return (
      <article className="feed-event feed_news_event_cal_4">
        <header>
          <h2>{this.props.heading}</h2>
        </header>
        <p className="text-muted">
          <b>KEY QUOTE</b>
        </p>
        <blockquote>
          <p>{this.props.top_sentence}</p>
          <small>
            <cite title={this.props.top_sender}>{this.props.top_sender}</cite>
          </small>
        </blockquote>
        <div className="text-right">
          <a className="text-muted">VIEW ALERTS</a>
        </div>
      </article>
    );
  }
}

ReactDOM.render(
  <Feed />,
  document.getElementById('root')
);
