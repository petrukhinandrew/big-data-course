import * as React from "react"
import * as ReactDOM from "react-dom"
import { useState } from 'react'


function Button() {
  const [title, setTitle] = useState("<placeholder>")
  const handleClick = function (event: React.MouseEvent<HTMLButtonElement, MouseEvent>) {
    console.log(event)
    setTitle(prompt("enter new title", "no data"))
  }
  return (
  <>
  <h1>{title}</h1>
  <button onClick={handleClick} title="Click me">Click Me</button>
  </>)
}

ReactDOM.render(
  <React.StrictMode>
    <Button></Button>
  </React.StrictMode>,
  document.getElementById("root")
)