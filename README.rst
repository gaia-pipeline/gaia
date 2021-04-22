.. raw:: html

    <img src="https://cdn.rawgit.com/michelvocks/ef3894f63c3bb004bca1a2fd5f7eb644/raw/c36d614db8afe229b466b38de1636a82ad809f64/gaia-logo-text.png" width="650px">

|build-status| |go-report| |go-doc| |apache2| |chat| |codecov|

Gaia is an open source automation platform which makes it easy and fun to build powerful pipelines in any programming language. Based on `HashiCorp's go-plugin`_ and `gRPC`_, gaia is efficient, fast, lightweight, and developer friendly.

Develop powerful `pipelines <What is a pipeline?_>`_ with the help of `SDKs <Why do I need an SDK?_>`_ and simply check-in your code into a git repository. Gaia automatically clones your code repository, compiles your code to a binary, and executes it on-demand. All results are streamed back and formatted as a user-friendly graphical output.

Check out `gaia-pipeline.io`_ to learn more.

Motivation
==========

.. begin-motivation

*Automation Engineer*, *DevOps Engineer*, *SRE*, *Cloud Engineer*,
*Platform Engineer* - they all have one in common:
The majority of tech people are not motivated to take up this work and they are hard to recruit.

One of the main reasons for this is the abstraction and poor execution of many automation tools. They come with their own configuration (`YAML`_ syntax) specification or limit the user to one specific programming language. Testing is nearly impossible because most automation tools lack the ability to mock services and subsystems. Even tiny things, for example parsing a JSON file, are sometimes really painful because external, outdated libraries were used and not included in the standard framework.

We believe it's time to remove all those abstractions and come back to our roots. Are you tired of writing endless lines of YAML-code? Are you sick of spending days forced to write in a language that does not suit you and is not fun at all? Do you enjoy programming in a language you like? Then Gaia is for you.

How does it work?
=================

.. begin-architecture

Gaia is based on `HashiCorp's go-plugin`_. It's a `plugin system`_ that uses `gRPC`_ to communicate over `HTTP/2`_. Initially, HashiCorp developed this tool for `Packer`_ but now it's heavily used by `Terraform`_, `Nomad`_, and `Vault`_ too.

Plugins, also called `pipelines <What is a pipeline?_>`_, are applications which can be written in any programming language, as long as `gRPC`_ is supported. All functions, also called `jobs <What is a job?>`_, are exposed to Gaia and can form up a dependency graph that describes the order of execution.

Pipelines can be compiled locally or simply over the integrated build system. Gaia clones the git repository and automatically builds the included pipeline. If a change (`git push`_) happened, Gaia will automatically rebuild the pipeline for you*.

After a pipeline has been started, all log output is returned back to Gaia and displayed in a detailed overview with their final result status.

Gaia uses `boltDB` for storage. This makes the installation process super easy. No external database is currently required.

\* *This requires polling or webhook to be activated.*

Screenshots
===========

.. begin-screenshots

|sh-login|
|sh-overview|
|sh-create-pipeline|
|sh-pipeline-detailed|
|sh-pipeline-logs|
|sh-vault|
|sh-settings|

Getting Started
===============

.. begin-getting-started

Installation
------------

The installation of gaia is simple and often takes a few minutes.

Using docker
~~~~~~~~~~~~

The following command starts gaia as a daemon process and mounts all data to the current folder. Afterwards, gaia will be available on the host system on port 8080. Use the standard user **admin** and password **admin** as initial login. It is recommended to change the password afterwards.

.. code:: sh

    docker run -d -p 8080:8080 -v $PWD:/data gaiapipeline/gaia:latest

This uses the image with the *latest* tag which includes all required libraries and compilers for all supported languages. If you prefer a smaller image suited for your preferred language, have a look at the `available docker image tags`_.

Manually
~~~~~~~~

It is possible to install Gaia directly on the host system.
This can be achieved by downloading the binary from the `releases page`_.

Gaia will automatically detect the folder of the binary and will place all data next to it. You can change the data directory with the startup parameter *-home-path* if you want.

Using helm
~~~~~~~~~~

If you haven't got an ingress controller pod yet, make sure that you have `kube-dns` or `coredns` enabled, run this command to set it up.

.. code:: sh

    make kube-ingress

To init helm:

.. code:: sh

    helm init

To deploy gaia:

.. code:: sh

    make deploy-kube

Example Pipelines
-----

Go
~~~

.. code:: go

    package main

    import (
        "log"

	sdk "github.com/gaia-pipeline/gosdk"
    )

    // This is one job. Add more if you want.
    func DoSomethingAwesome(args sdk.Arguments) error {
        log.Println("This output will be streamed back to gaia and will be displayed in the pipeline logs.")

	// An error occurred? Return it back so gaia knows that this job failed.
	return nil
    }

    func main() {
        jobs := sdk.Jobs{
            sdk.Job{
                Handler:     DoSomethingAwesome,
	        Title:       "DoSomethingAwesome",
		Description: "This job does something awesome.",
	    },
	}

	// Serve
	if err := sdk.Serve(jobs); err != nil {
	    panic(err)
	}
}

Python
~~~~~~~

.. code:: python

    from gaiasdk import sdk
    import logging

    def MyAwesomeJob(args):
        logging.info("This output will be streamed back to gaia and will be displayed in the pipeline logs.")
        # Just raise an exception to tell Gaia if a job failed.
        # raise Exception("Oh no, this job failed!")

    def main():
        logging.basicConfig(level=logging.INFO)
        myjob = sdk.Job("MyAwesomeJob", "Do something awesome", MyAwesomeJob)
        sdk.serve([myjob])

Java
~~~~

.. code:: java

    package io.gaiapipeline;

    import io.gaiapipeline.javasdk.*;

    import java.util.ArrayList;
    import java.util.Arrays;
    import java.util.logging.Logger;

    public class Pipeline
    {
        private static final Logger LOGGER = Logger.getLogger(Pipeline.class.getName());

        private static Handler MyAwesomeJob = (gaiaArgs) -> {
            LOGGER.info("This output will be streamed back to gaia and will be displayed in the pipeline logs.");
	    // Just raise an exception to tell Gaia if a job failed.
            // throw new IllegalArgumentException("Oh no, this job failed!");
        };

        public static void main( String[] args )
        {
            PipelineJob myjob = new PipelineJob();
            myjob.setTitle("MyAwesomeJob");
            myjob.setDescription("Do something awesome.");
            myjob.setHandler(MyAwesomeJob);

            Javasdk sdk = new Javasdk();
            try {
                sdk.Serve(new ArrayList<>(Arrays.asList(myjob)));
            } catch (Exception ex) {
                ex.printStackTrace();
            }
        }
    }

C++
~~~~

.. code:: cpp

   #include "cppsdk/sdk.h"
   #include <list>
   #include <iostream>

   void DoSomethingAwesome(std::list<gaia::argument> args) throw(std::string) {
      std::cerr << "This output will be streamed back to gaia and will be displayed in the pipeline logs." << std::endl;

      // An error occurred? Return it back so gaia knows that this job failed.
      // throw "Uhh something badly happened!"
   }

   int main() {
      std::list<gaia::job> jobs;
      gaia::job awesomejob;
      awesomejob.handler = &DoSomethingAwesome;
      awesomejob.title = "DoSomethingAwesome";
      awesomejob.description = "This job does something awesome.";
      jobs.push_back(awesomejob);

      try {
         gaia::Serve(jobs);
      } catch (string e) {
         std::cerr << "Error: " << e << std::endl;
      }
   }

Ruby
~~~~

.. code:: ruby

   require 'rubysdk'

   class Main
       AwesomeJob = lambda do |args|
           STDERR.puts "This output will be streamed back to gaia and will be displayed in the pipeline logs."

           # An error occurred? Raise an exception and gaia will fail the pipeline.
           # raise "Oh gosh! Something went wrong!"
       end

       def self.main
           awesomejob = Interface::Job.new(title: "Awesome Job",
                                           handler: AwesomeJob,
                                           desc: "This job does something awesome.")

           begin
               RubySDK.Serve([awesomejob])
           rescue => e
               puts "Error occured: #{e}"
               exit(false)
           end
       end
   end

Node.JS
~~~~

.. code:: javascript

   const nodesdk = require('@gaia-pipeline/nodesdk');

   function DoSomethingAwesome(args) {
       console.error('This output will be streamed back to gaia and will be displayed in the pipeline logs.');

       // An error occurred? Throw it back so gaia knows that this job failed.
       // throw new Error('My error message');
   }

   // Serve
   try {
       nodesdk.Serve([{
           handler: DoSomethingAwesome,
           title: 'DoSomethingAwesome',
           description: 'This job does something awesome.'
       }]);
   } catch (err) {
       console.error(err);
   }

Pipelines are defined by jobs and a function usually represents a job. You can define as many jobs in your pipeline as you want.

Every function accepts arguments. Those arguments can be requested from the pipeline itself and the values are passed back in from the UI.

Some pipeline jobs need a specific order of execution. `DependsOn` allows you to declare dependencies for every job.

You can find real examples and more information on `how to develop a pipeline`_ in the docs.

Security
========

See the Documentation located here: `security-docs`_.

Documentation and more
======================

Please find the docs at https://docs.gaia-pipeline.io. We also have a tutorials section over there with examples and real use-case scenarios. For example, `Kubernetes deployment with vault integration`_.

Questions and Answers (Q&A)
---------------------------

What problem solves **Gaia**?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
Literally every tool that was designed for automation, continuous integration (CI), and continuous deployment (CD) like Spinnaker, Jenkins, Gitlab CI/CD, TravisCI, CircleCI, Codeship, Bamboo and many more, introduced their own configuration format. Some of them don't even support *configuration/automation as code*. This works well for simple tasks like running a ``go install`` or ``mvn clean install`` but in the real world there is more to do.

Gaia is the first platform that does not limit the user and provides full support for almost all common programming languages without losing the features offered by todays CI/CD tools.

What is a **pipeline**?
~~~~~~~~~~~~~~~~~~~~~~~
A pipeline is a real application with at least one function (we call it a Job). Every programming language can be used as long as gRPC is supported. We offer SDKs to support the development.

What is a **job**?
~~~~~~~~~~~~~~~~~~
A job is a function, usually globally exposed to Gaia. Dependent on the dependency graph, Gaia will execute this function in a specific order.

Why do I need an **SDK**?
~~~~~~~~~~~~~~~~~~~~~~~~~~
The SDK implements the Gaia plugin gRPC interface and offers helper functions like serving the gRPC-Server. This helps you to focus on the real problem instead of doing the boring stuff.

Which programming languages are supported?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
We currently fully support Go, Java, Python, C++, Ruby and Node.JS.

When do you support programming language **XYZ**?
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
We are working hard to support as much programming languages as possible but our resources are limited and we are also mostly no experts in all programming languages. If you are willing to contribute, feel free to open an issue and start working.

Roadmap
=======

Gaia is currently available as beta version.

Feel free to open a new GitHub issue to request a new feature.

Contributing
============

Gaia can only evolve and become a great product with the help of contributors. If you like to contribute, please have a look at our `issues section`_. We do our best to mark issues for new contributors with the label *good first issue*.

If you think you found a good first issue, please consider this list as a short guide:

* If the issue is clear and you have no questions, please leave a short comment that you started working on this. The issue will be usually blocked for two weeks for you to solve it.
* If something is not clear or you are unsure what to do, please leave a comment so we can add more detailed description.
* Make sure your development environment is configured and set up. You need `Go installed`_ on your machine and also `nodeJS`_ for the frontend. Clone this repository and run the **make** command inside the cloned folder. This will start the backend. To start the frontend you have to open a new terminal window and go into the frontend folder. There you run **npm install** and then **npm run serve**. This should automatically open a new browser window.
* Before you start your work, you should fork this repository and push changes to your fork. Afterwards, send a merge request back to upstream.

Contact
=======

If you have any questions feel free to contact us on `slack`_.

.. _`HashiCorp's go-plugin`: https://github.com/hashicorp/go-plugin
.. _`gRPC`: https://grpc.io/
.. _`Do not use it for mission critical jobs yet!`: https://tenor.com/view/enter-at-your-own-risk-gif-8912210
.. _`YAML`: https://en.wikipedia.org/wiki/YAML
.. _`releases page`: https://github.com/gaia-pipeline/gaia/releases
.. _`Packer`: https://www.packer.io/
.. _`Terraform`: https://www.terraform.io/
.. _`Nomad`: https://www.nomadproject.io/
.. _`Vault`: https://www.vaultproject.io/
.. _`boltDB`: https://github.com/coreos/bbolt
.. _`Unix nice level`: https://en.wikipedia.org/wiki/Nice_(Unix)
.. _`issues section`: https://github.com/gaia-pipeline/gaia/issues
.. _`Go installed`: https://golang.org/doc/install
.. _`nodeJS`: https://nodejs.org/
.. _`go-example repo`: https://github.com/gaia-pipeline/go-example
.. _`slack`: https://slack.gaia-pipeline.io/
.. _`Kubernetes deployment with vault integration`: https://docs.gaia-pipeline.io/tutorials/kube-vault-deploy/
.. _`git push`: https://git-scm.com/docs/git-push
.. _`HTTP/2`: https://http2.github.io/
.. _`security-docs`: https://github.com/gaia-pipeline/gaia/blob/master/security/README.md
.. _`plugin system`: https://en.wikipedia.org/wiki/Plug-in_(computing)
.. _`available docker image tags`: https://hub.docker.com/r/gaiapipeline/gaia/tags/
.. _`how to develop a pipeline`: https://docs.gaia-pipeline.io/develop-pipelines/
.. _`gaia-pipeline.io`: https://gaia-pipeline.io/

.. |build-status| image:: https://circleci.com/gh/gaia-pipeline/gaia/tree/master.svg?style=shield&circle-token=c0e15edfb08f8076076cbbb55558af6cfecb89b8
    :alt: Build Status
    :scale: 100%
    :target: https://circleci.com/gh/gaia-pipeline/gaia/tree/master

.. |go-report| image:: https://goreportcard.com/badge/github.com/gaia-pipeline/gaia
    :alt: Go Report Card
    :target: https://goreportcard.com/report/github.com/gaia-pipeline/gaia

.. |go-doc| image:: https://godoc.org/github.com/gaia-pipeline/gaia?status.svg
    :alt: GoDoc
    :target: https://godoc.org/github.com/gaia-pipeline/gaia

.. |apache2| image:: https://img.shields.io/badge/license-Apache-blue.svg
    :alt: Apache licensed
    :target: https://github.com/gaia-pipeline/gaia/blob/master/LICENSE

.. |chat| image:: https://gaia-slack-invite.herokuapp.com/badge.svg
    :alt: Slack
    :target: https://gaia-slack-invite.herokuapp.com/

.. |codecov| image:: https://codecov.io/gh/gaia-pipeline/gaia/branch/master/graph/badge.svg
    :target: https://codecov.io/gh/gaia-pipeline/gaia

.. |sh-login| image:: https://github.com/gaia-pipeline/gaia/blob/master/screenshots/login.png
    :alt: gaia login screenshot
    :width: 650px

.. |sh-overview| image:: https://github.com/gaia-pipeline/gaia/blob/master/screenshots/overview.png
    :alt: gaia overview screenshot
    :width: 650px

.. |sh-create-pipeline| image:: https://github.com/gaia-pipeline/gaia/blob/master/screenshots/create-pipeline.png
    :alt: gaia create pipeline screenshot
    :width: 650px

.. |sh-vault| image:: https://github.com/gaia-pipeline/gaia/blob/master/screenshots/vault.png
    :alt: gaia Vault screenshot
    :width: 650px

.. |sh-pipeline-detailed| image:: https://github.com/gaia-pipeline/gaia/blob/master/screenshots/detail-pipeline.png
    :alt: gaia pipeline detailed screenshot
    :width: 650px

.. |sh-pipeline-logs| image:: https://github.com/gaia-pipeline/gaia/blob/master/screenshots/logs-pipeline.png
    :alt: gaia pipeline logs screenshot
    :width: 650px

.. |sh-settings| image:: https://github.com/gaia-pipeline/gaia/blob/master/screenshots/settings.png
    :alt: gaia settings screenshot
    :width: 650px

