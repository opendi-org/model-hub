---
title: |
  [OpenDI Model Hub]{.mark}

  User Manual

  [Core Features and Use Cases]{.mark}

  [Open DI]{.mark}

  CSC 492 Team 22

  []{#_hrz496vv3n2u .anchor}Connor Blumsack

  Matthew Bunch  
  Eric Jun

  Alex Mize  
  Jay Pham  
    

  North Carolina State University

  Department of Computer Science

  []{#_luwv8e5xheve .anchor}4/25/2025
---

# Introduction

A guide designed for end users and provides clear, step-by-step
instructions for interacting with the application.

# Table of Contents

[**Introduction 2**](#introduction)

[**Table of Contents 2**](#table-of-contents)

[**Core Features and Use Cases 3**](#core-features-and-use-cases)

> [Logging In 3](#logging-in)
>
> [Uploading Models 4](#uploading-models)
>
> [Viewing Models 5](#viewing-models)
>
> [Downloading Models 6](#downloading-models)
>
> [Updating Models 7](#updating-models)
>
> [Searching for Models 10](#searching-for-models)
>
> [Viewing History of Models 11](#viewing-history-of-models)
>
> [Viewing Fork Information of Models
> 12](#viewing-fork-information-of-models)

# 

#   {#section-1}

# Core Features and Use Cases

## Logging In

If the user clicks on the *Login* button in the top right corner of the
navigation bar, then they can sign in using their email address and
password.

**Note:** New users will be automatically created upon logging in for
the first time.

![](media/image5.png){width="6.5in" height="4.5in"}

![](media/image16.png){width="6.5in" height="2.4722222222222223in"}

![](media/image21.png){width="6.5in" height="3.6666666666666665in"}

## Uploading Models

If the user clicks the *Upload* button in the navigation bar, then they
will be directed to the model upload detail page. The user can either
drag and drop a .json model file or click the dotted box under **My
Assets** to select a .json file from the file
explorer.![](media/image23.png){width="6.5in"
height="3.7239588801399823in"}

![](media/image2.png){width="6.5in" height="3.0694444444444446in"}

![](media/image18.png){width="6.5in" height="2.1805555555555554in"}

## Viewing Models

If the user clicks the *View* button for any of the models on the home
page, then they will be directed to the model detail page.

![](media/image17.png){width="6.5in" height="4.0in"}

![](media/image8.png){width="6.5in" height="2.7222222222222223in"}

## Downloading Models

If the user navigates to a model detail page and clicks the Download
button, they will be prompted to choose a location in their file system
to save the model.

![](media/image12.png){width="6.5in" height="2.6944444444444446in"}

![](media/image6.png){width="6.5in" height="3.0416666666666665in"}

## Updating Models

If the user navigates to a model detail page and clicks the *Upload*
button, they will be prompted to upload a .json file to update the
current model. It is important to note that in this current state if you
update a component of a model that is potentially used by other models
it will be rolled back unless you also update the UUID for this
component.

![](media/image22.png){width="6.5in" height="2.9305555555555554in"}

![](media/image14.png){width="6.5in" height="2.8333333333333335in"}

![](media/image3.png){width="1.96875in"
height="0.6875in"}![](media/image19.png){width="6.5in"
height="2.7777777777777777in"}  
![](media/image20.png){width="6.5in" height="3.9444444444444446in"}

## Searching for Models

If the user clicks the *Search* button in the navigation bar, they will
be directed to a search page where they can look for models and filter
results by model name or creator name.

Alternatively, if the user searches for a model directly using the
search bar in the navigation bar, they will be directed to the search
page with the results already populated.

![](media/image11.png){width="6.5in" height="3.763888888888889in"}

![](media/image15.png){width="6.5in" height="1.625in"}

Alternatively...

![](media/image13.png){width="6.5in" height="1.7083333333333333in"}

![](media/image4.png){width="6.5in" height="1.7083333333333333in"}

## Viewing History of Models

If the user navigates to a model detail page and clicks the *COMMIT
DIFF* tab, they will see a drop down list that will let them see the
version history of any model.

![](media/image1.png){width="6.5in" height="4.236111111111111in"}

## Viewing Fork Information of Models

If the user navigates to a model detail page and clicks the *FORK INFO*
tab, they will see the model\'s parent lineage as well as any child
models it may have. Clicking on any of the listed models will direct the
user to that model's detail page.

![](media/image9.png){width="6.5in" height="2.986111111111111in"}

![](media/image7.png){width="6.5in" height="3.263888888888889in"}

![](media/image10.png){width="6.5in" height="2.7777777777777777in"}

#  {#section-2}
