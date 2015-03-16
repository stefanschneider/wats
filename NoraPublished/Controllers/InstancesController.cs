﻿using System;
using System.Web.Http;

namespace nora.Controllers
{
    public class InstancesController : ApiController
    {
        [Route("~/")]
        [HttpGet]
        public IHttpActionResult Root()
        {
            return Ok("hello i am nora");
        }

        [Route("~/id")]
        [HttpGet]
        public IHttpActionResult Id()
        {
            const string uuid = "A123F285-26B4-45F1-8C31-816DC5F53ECF";
            return Ok(uuid);
        }

        [Route("~/env")]
        [HttpGet]
        public IHttpActionResult Env()
        {
            return Ok(Environment.GetEnvironmentVariables());
        }

        [Route("~/env/:name")]
        [HttpGet]
        public IHttpActionResult EnvName(string name)
        {
            return Ok(Environment.GetEnvironmentVariable(name));
        }

        [Route("~/crash")]
        public IHttpActionResult Crash()
        {
            for (var i = 0; i <= 40; i++)
                Console.WriteLine(i);

            Environment.Exit(123);

            return Ok("Should never be reached");
        }
    }
}